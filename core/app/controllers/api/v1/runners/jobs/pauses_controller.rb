class Api::V1::Runners::Jobs::PausesController < Api::V1::BaseController
  before_action :set_job

  def create
    @job.pause!
    render partial: "api/v1/runners/jobs/job", locals: { job: @job }
  rescue ActiveRecord::RecordInvalid => error
    render_json_error(
      status: :unprocessable_entity,
      message: error.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end

  def destroy
    @job.resume!
    render partial: "api/v1/runners/jobs/job", locals: { job: @job }
  rescue ActiveRecord::RecordInvalid => error
    render_json_error(
      status: :unprocessable_entity,
      message: error.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end

  private
    def set_job
      runner = Current.account.runners.find(params[:runner_id])
      @job = runner.jobs.find(params[:job_id])
    end
end
