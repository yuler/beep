class Api::V1::Runner::Jobs::SyncsController < Api::V1::Runner::BaseController
  def create
    jobs_payload = Array(params[:jobs])
    @synced_jobs = []

    ActiveRecord::Base.transaction do
      jobs_payload.each do |job_data|
        attrs = job_attrs_from(job_data)
        next if attrs[:slug].blank?

        job = find_or_build_job(id: job_data[:id], slug: attrs[:slug])
        job.assign_attributes(attrs)
        job.save!
        @synced_jobs << job
      end
    end

    render template: "api/v1/runner/jobs/syncs/create"
  rescue ActiveRecord::RecordInvalid => e
    render_json_error(
      status: :unprocessable_entity,
      message: e.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end

  private

    def find_or_build_job(id:, slug:)
      if id.present? && (job = @current_runner.jobs.find_by(id: id))
        job
      else
        @current_runner.jobs.find_or_initialize_by(slug: slug)
      end
    end

    def job_attrs_from(data)
      slug = data[:slug].to_s.strip.downcase
      name = data[:name].presence || (slug.present? ? slug.tr("_-", " ").titleize : nil)
      cron = data.key?(:cron) ? data[:cron].presence : "*/5 * * * *"
      timezone = IanaTimezone.resolve(data[:timezone])
      timeout_seconds = data[:timeout_seconds].presence || 30
      config = data[:config].respond_to?(:to_unsafe_h) ? data[:config].to_unsafe_h : (data[:config] || {})
      if data[:description].present? && !config.key?("description")
        config = config.merge("description" => data[:description].to_s)
      end

      {
        account: @current_runner.account,
        slug: slug,
        name: name,
        cron: cron,
        timezone: timezone,
        timeout_seconds: timeout_seconds,
        config: config
      }
    end
end
