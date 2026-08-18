class Api::V1::SettingsController < Api::V1::BaseController
  def show
    @account = Current.account
    render :show
  end

  def update
    @account = Current.account
    unless @account.personal?
      return render_json_error(
        status: :unprocessable_entity,
        message: "Email reminders are only available on personal accounts",
        code: "VALIDATION_ERROR"
      )
    end

    if @account.update(settings_params)
      render :show
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @account.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  private
    def settings_params
      params.permit(:email_channel_enabled)
    end
end
