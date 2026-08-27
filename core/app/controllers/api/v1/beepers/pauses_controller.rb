class Api::V1::Beepers::PausesController < Api::V1::BaseController
  before_action :set_beeper

  def create
    @beeper.pause!
    render partial: "api/v1/beepers/beeper", locals: { beeper: @beeper }
  rescue ActiveRecord::RecordInvalid => error
    render_json_error(
      status: :unprocessable_entity,
      message: error.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end

  def destroy
    @beeper.resume!
    render partial: "api/v1/beepers/beeper", locals: { beeper: @beeper }
  rescue ActiveRecord::RecordInvalid => error
    render_json_error(
      status: :unprocessable_entity,
      message: error.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end

  private

  def set_beeper
    @beeper = Current.account.beepers.find(params[:beeper_id])
  end
end
