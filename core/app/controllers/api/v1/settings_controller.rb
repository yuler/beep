class Api::V1::SettingsController < Api::V1::BaseController
  def show
    @account = Current.account
    render :show
  end

  def update
    @account = Current.account

    if Current.user.update(settings_params)
      render :show
    else
      render_json_error(
        status: :unprocessable_entity,
        message: Current.user.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  private
    def settings_params
      params.permit(notification_channels: [])
    end
end
