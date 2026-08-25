class Api::V1::Beeps::PausesController < Api::V1::BaseController
  before_action :set_beep

  def create
    @beep.pause!
    render partial: "api/v1/beeps/beep", locals: { beep: @beep }
  rescue ActiveRecord::RecordInvalid => error
    render_json_error(
      status: :unprocessable_entity,
      message: error.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end

  def destroy
    @beep.resume!
    render partial: "api/v1/beeps/beep", locals: { beep: @beep }
  rescue ActiveRecord::RecordInvalid => error
    render_json_error(
      status: :unprocessable_entity,
      message: error.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end

  private
    def set_beep
      @beep = Current.account.beeps.find(params[:beep_id])
    end
end
