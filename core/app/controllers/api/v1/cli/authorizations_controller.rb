class Api::V1::Cli::AuthorizationsController < Api::V1::BaseController
  disallow_account_scope
  allow_unauthenticated_access only: %i[ create ]
  rate_limit to: 20, within: 1.minute, only: %i[ create ],
    by: -> { request.remote_ip }, with: :rate_limit_exceeded
  rate_limit to: 30, within: 1.minute, only: %i[ show update destroy ],
    by: -> { request.remote_ip },
    with: :rate_limit_exceeded

  def create
    @auth = Cli::Authorization.create_request!(client_name: params[:client_name])
    @web_origin = Rails.configuration.x.web_origin
    render :create, status: :created
  end

  def show
    @auth = Cli::Authorization.active.find_by(user_code: params[:user_code].to_s.upcase.strip)
    if @auth
      render :show, status: :ok
    else
      render_json_error(status: :not_found, message: "Invalid or expired user code", code: "NOT_FOUND")
    end
  end

  def update
    @auth = Cli::Authorization.active.find_by(user_code: params[:user_code].to_s.upcase.strip)
    identity = Current.identity
    if @auth && identity && @auth.approve!(identity: identity, client_name: params[:client_name])
      @access_token = @auth.access_token
      render :update, status: :ok
    else
      code = @auth&.errors&.present? ? "VALIDATION_ERROR" : "UNPROCESSABLE_ENTITY"
      message = @auth&.errors&.full_messages&.to_sentence.presence || "Unable to approve authorization"
      render_json_error(status: :unprocessable_entity, message: message, code: code)
    end
  end

  def destroy
    @auth = Cli::Authorization.active.find_by(user_code: params[:user_code].to_s.upcase.strip)
    if @auth&.deny!
      head :no_content
    else
      render_json_error(status: :unprocessable_entity, message: "Unable to deny authorization", code: "UNPROCESSABLE_ENTITY")
    end
  end

  private
    def rate_limit_exceeded
      @device_flow_error = "slow_down"
      @device_flow_error_description = "Too many requests"
      render "api/v1/cli/authorizations/device_flow_error", status: :too_many_requests
    end
end
