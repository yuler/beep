class Api::V1::WebPushController < Api::V1::BaseController
  disallow_account_scope

  def show
    if public_key = Rails.application.config.x.vapid.public_key
      @vapid_public_key = public_key
      render :show
    else
      render_json_error(
        status: :service_unavailable,
        message: "Web Push is not configured",
        code: "WEB_PUSH_UNAVAILABLE"
      )
    end
  end
end
