class Api::V1::Runner::JobsController < Api::V1::Runner::BaseController
  def index
    @jobs = @current_runner.jobs.order(:name)
    render :index
  end

  def create
    slug = params[:slug].to_s.strip.downcase
    name = params[:name].presence || slug.tr("_-", " ").titleize
    cron = params[:cron].presence || "*/5 * * * *"
    timezone = IanaTimezone.resolve(params[:timezone])
    timeout_seconds = params[:timeout_seconds].presence || 30
    config = params[:config].respond_to?(:to_unsafe_h) ? params[:config].to_unsafe_h : (params[:config] || {})

    @job = @current_runner.jobs.find_or_initialize_by(slug: slug)
    @job.assign_attributes(
      account: @current_runner.account,
      name: name,
      cron: cron,
      timezone: timezone,
      timeout_seconds: timeout_seconds,
      config: config
    )

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
        slug = job_data[:slug].to_s.strip.downcase
        next if slug.blank?

        name = job_data[:name].presence || slug.tr("_-", " ").titleize
        cron = job_data[:cron].presence || "*/5 * * * *"
        timezone = IanaTimezone.resolve(job_data[:timezone])
        timeout_seconds = job_data[:timeout_seconds].presence || 30
        config = job_data[:config].respond_to?(:to_unsafe_h) ? job_data[:config].to_unsafe_h : (job_data[:config] || {})

        job = @current_runner.jobs.find_or_initialize_by(slug: slug)
        job.assign_attributes(
          account: @current_runner.account,
          name: name,
          cron: cron,
          timezone: timezone,
          timeout_seconds: timeout_seconds,
          config: config
        )
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
end
