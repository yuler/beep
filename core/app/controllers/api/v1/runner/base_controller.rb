class Api::V1::Runner::BaseController < ActionController::API
  include ActionController::Cookies
  include Api::V1::Responses

  before_action :authenticate_runner!

  private

  def authenticate_runner!
    token = extract_runner_token
    if token.blank?
      render_json_error(
        status: :unauthorized,
        message: "Missing runner token",
        code: "UNAUTHORIZED"
      )
      return
    end

    @current_runner = Runner.find_by_raw_token(token)
    unless @current_runner
      render_json_error(
        status: :unauthorized,
        message: "Invalid runner token",
        code: "UNAUTHORIZED"
      )
    end
  end

  def extract_runner_token
    token = request.headers["X-Runner-Token"].presence
    return token if token.present?

    auth_header = request.headers["Authorization"].presence
    if auth_header&.start_with?("Bearer ")
      auth_header.delete_prefix("Bearer ").strip
    end
  end
end
