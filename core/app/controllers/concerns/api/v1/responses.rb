module Api::V1::Responses
  extend ActiveSupport::Concern

  def render_json_ok
    render json: { message: "OK" }, status: :ok
  end

  def render_json_created(json: {})
    render json: json, status: :created
  end

  def render_json_unauthorized
    render_json_error(status: :unauthorized, message: "Unauthorized", code: "UNAUTHORIZED")
  end

  def render_json_not_found
    render_json_error(status: :not_found, message: "Not Found", code: "NOT_FOUND")
  end

  def render_json_too_many_requests
    render_json_error(status: :too_many_requests, message: "Too Many Requests", code: "TOO_MANY_REQUESTS")
  end

  def render_json_error(status:, message:, code: nil)
    render json: { code:, message: }, status:
  end

  def render_json(json: {}, status: :ok)
    render json: json, status: status
  end
end
