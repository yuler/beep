class Api::V1::BeeperInstalls::PausesController < Api::V1::BaseController
  before_action :set_beeper_install

  def create
    @beeper_install.pause!
    render partial: "api/v1/beeper_installs/beeper_install", locals: { beeper_install: @beeper_install }
  rescue ActiveRecord::RecordInvalid => error
    render_json_error(
      status: :unprocessable_entity,
      message: error.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end

  def destroy
    @beeper_install.resume!
    render partial: "api/v1/beeper_installs/beeper_install", locals: { beeper_install: @beeper_install }
  rescue ActiveRecord::RecordInvalid => error
    render_json_error(
      status: :unprocessable_entity,
      message: error.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end

  private

  def set_beeper_install
    @beeper_install = Current.account.beeper_installs.find(params[:beeper_install_id])
  end
end
