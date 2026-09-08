class Api::V1::Runners::RunsController < Api::V1::BaseController
  before_action :set_job

  def index
    @runs = @job.runs.order(scheduled_for: :desc).limit(50)
    render :index
  end

  def show
    @run = @job.runs.find(params[:id])
    render :show
  end

  def create
    @run = @job.trigger_run!
    render :show, status: :created
  end

  def destroy
    run = deletable_run
    return if performed?

    run.destroy!
    head :no_content
  end

  def clear
    @job.runs.where(status: terminal_statuses).delete_all
    head :no_content
  end

  private
    def set_job
      runner = Current.account.runners.find(params[:runner_id])
      @job = runner.jobs.find(params[:job_id])
    end

    def deletable_run
      run = @job.runs.find(params[:id])
      unless run.status.in?(terminal_statuses)
        render_json_error(
          status: :unprocessable_entity,
          message: "Run is still pending or running",
          code: "VALIDATION_ERROR"
        )
        return
      end
      run
    end

    def terminal_statuses
      %w[ succeeded failed expired ]
    end
end
