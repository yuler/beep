class Api::V1::BeepsController < Api::V1::BaseController
  def index
    render_json json: {
      beeps: Current.account.beeps.order(created_at: :desc).map { |beep| beep_json(beep) }
    }
  end

  def create
    beep = Current.account.beeps.new(beep_params.merge(kind: :once))

    if beep.save
      render_json_created json: beep_json(beep)
    else
      render_json_error(
        status: :unprocessable_entity,
        message: beep.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  private
    def beep_params
      params.permit(:message, :run_at)
    end

    def beep_json(beep)
      {
        id: beep.id,
        message: beep.message,
        kind: beep.kind,
        status: beep.status,
        run_at: beep.run_at,
        next_run_at: beep.next_run_at,
        timezone: beep.timezone
      }
    end
end
