class Api::V1::Runners::Runs::HistoriesController < Api::V1::BaseController
  before_action :set_job

  def destroy
    @job.runs.terminal.delete_all
    head :no_content
  end

  private
    def set_job
      runner = Current.account.runners.find(params[:runner_id])
      @job = runner.jobs.find(params[:job_id])
    end
end
