class Api::V1::RunnerJobRunsController < Api::V1::BaseController
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

  private

  def set_job
    runner = Current.account.runners.find(params[:runner_id])
    @job = runner.jobs.find(params[:job_id])
  end
end
