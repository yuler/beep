class Api::V1::RunnerJobsController < Api::V1::BaseController
  before_action :set_runner
  before_action :set_job, only: %i[ show update destroy ]

  def index
    @jobs = @runner.jobs.order(created_at: :desc)
    render :index
  end

  def show
    render :show
  end

  def create
    attrs = job_params
    attrs[:account] = Current.account
    attrs[:timezone] = job_timezone
    if params[:description].present?
      config = (attrs[:config] || {}).dup
      config["description"] = params[:description].to_s
      attrs[:config] = config
    end

    @job = @runner.jobs.new(attrs)

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

  def update
    attrs = update_params
    attrs[:timezone] = IanaTimezone.resolve(params[:timezone]) if params.key?(:timezone)
    if params.key?(:description)
      config = (attrs[:config] || @job.config || {}).dup
      config["description"] = params[:description].to_s
      attrs[:config] = config
    end

    if @job.update(attrs)
      render :show
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @job.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  def destroy
    @job.destroy!
    head :no_content
  end

  private

    def set_runner
      @runner = Current.account.runners.find(params[:runner_id])
    end

    def set_job
      @job = @runner.jobs.find(params[:id])
    end

    def job_params
      attrs = params.permit(:name, :slug, :cron, :timeout_seconds)
      attrs[:config] = params[:config].to_unsafe_h if params[:config].respond_to?(:to_unsafe_h)
      attrs
    end

    def update_params
      attrs = params.permit(:name, :slug, :cron, :timeout_seconds)
      attrs[:config] = params[:config].to_unsafe_h if params[:config].respond_to?(:to_unsafe_h)
      attrs
    end

    def job_timezone
      IanaTimezone.resolve(Current.user.timezone, params[:timezone])
    end
end
