class Api::V1::Device::BaseController < ActionController::API
  include ActionController::Cookies
  include Api::V1::Responses

  before_action :authenticate_device!

  private
    def authenticate_device!
      token = extract_device_token
      if token.blank?
        render_json_error(
          status: :unauthorized,
          message: "Missing device token",
          code: "UNAUTHORIZED"
        )
        return
      end

      @current_device = Channel.active.device.find_by(token: token)
      unless @current_device
        render_json_error(
          status: :unauthorized,
          message: "Invalid device token",
          code: "UNAUTHORIZED"
        )
        return
      end

      @current_device.touch_last_seen
    end

    def extract_device_token
      request.headers["X-Device-Token"].to_s.strip.presence ||
        bearer_token
    end

    def bearer_token
      pattern = /^Bearer /
      header  = request.headers["Authorization"]
      header.gsub(pattern, "").strip if header && header.match(pattern)
    end
end
