class Api::V1::BeepsController < Api::V1::BaseController
  def index
    @beeps = Current.account.beeps.order(created_at: :desc)
    render :index
  end

  def create
    @beep = Current.account.beeps.new(beep_params.merge(kind: :once))

    if @beep.save
      render :create, status: :created
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @beep.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  private
    def beep_params
      params.permit(:message, :run_at)
    end
end
