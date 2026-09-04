class Api::V1::Runner::JobsController < Api::V1::Runner::BaseController
  def index
    @jobs = @current_runner.jobs.order(:name)
    render :index
  end

  def create
    attrs = job_attrs_from(params)

    @job = find_or_build_job(id: params[:id], slug: attrs[:slug])
    @job.assign_attributes(attrs)

    if @job.save
      render :show, status: :created
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @job.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  def sync
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

    render :sync
  rescue ActiveRecord::RecordInvalid => e
    render_json_error(
      status: :unprocessable_entity,
      message: e.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end

  def destroy
    slug = params[:id].to_s.strip.downcase
    @job = @current_runner.jobs.find_by(slug: slug) || @current_runner.jobs.find_by(id: params[:id])

    if @job
      @job.destroy
      head :no_content
    else
      render_json_error(
        status: :not_found,
        message: "Job not found",
        code: "NOT_FOUND"
      )
    end
  end

  private

  # Prefer @id when present so a local filename rename updates the same
  # RunnerJob (slug change) instead of inserting a duplicate under the new slug.
  def find_or_build_job(id:, slug:)
    if id.present?
      existing = @current_runner.jobs.find_by(id: id)
      return existing if existing
    end

    @current_runner.jobs.find_or_initialize_by(slug: slug)
  end

  def job_attrs_from(data)
    slug = data[:slug].to_s.strip.downcase
    name = data[:name].presence || (slug.present? ? slug.tr("_-", " ").titleize : nil)
    cron = data[:cron].presence || "*/5 * * * *"
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
