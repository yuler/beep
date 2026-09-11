class Api::V1::Cli::BaseController < ActionController::API
  include ActionController::Cookies
  include Api::V1::Responses

  before_action :authenticate_cli_channel!

  private
    def authenticate_cli_channel!
      token = extract_cli_token
      if token.blank?
        render_json_error(
          status: :unauthorized,
          message: "Missing CLI token",
          code: "UNAUTHORIZED"
        )
        return
      end

      @current_channel = Channel.active.cli.find_by(token: token)
      unless @current_channel
        render_json_error(
          status: :unauthorized,
          message: "Invalid CLI token",
          code: "UNAUTHORIZED"
        )
        return
      end

      @current_channel.touch_last_seen
    end

    def extract_cli_token
      request.headers["X-CLI-Token"].to_s.strip.presence ||
        bearer_token
    end

    def bearer_token
      pattern = /^Bearer /
      header  = request.headers["Authorization"]
      header.gsub(pattern, "").strip if header && header.match(pattern)
    end
end
