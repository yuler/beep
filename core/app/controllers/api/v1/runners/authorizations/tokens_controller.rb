class Api::V1::Runners::Authorizations::TokensController < Api::V1::BaseController
  disallow_account_scope
  allow_unauthenticated_access
  rate_limit to: 30, within: 1.minute, only: :create,
    by: -> { params[:device_code].to_s.strip.presence || request.remote_ip },
    with: :rate_limit_exceeded

  def create
    unless params[:grant_type] == "urn:ietf:params:oauth:grant-type:device_code"
      render json: { error: "unsupported_grant_type", error_description: "grant_type must be urn:ietf:params:oauth:grant-type:device_code" }, status: :bad_request
      return
    end

    @auth = Runner::Authorization.find_by(device_code: params[:device_code].to_s.strip)
    unless @auth
      render json: { error: "invalid_grant", error_description: "Invalid device code" }, status: :bad_request
      return
    end

    if @auth.poll_interval_exceeded?
      render json: { error: "slow_down", error_description: "Polling too frequently" }, status: :bad_request
      return
    end

    @auth.poll!

    if @auth.status == "approved"
      @runner = @auth.consume_token!
      if @runner
        render :create, status: :ok
      else
        render json: { error: "invalid_grant", error_description: "Invalid device code" }, status: :bad_request
      end
    elsif @auth.status == "consumed"
      render json: { error: "invalid_grant", error_description: "Invalid device code" }, status: :bad_request
    elsif @auth.status == "access_denied"
      render json: { error: "access_denied", error_description: "The user denied the authorization request" }, status: :bad_request
    elsif @auth.expired? || @auth.status == "expired"
      render json: { error: "expired_token", error_description: "The device code has expired" }, status: :bad_request
    else
      render json: { error: "authorization_pending", error_description: "The authorization request is still pending" }, status: :bad_request
    end
  end

  private
    def rate_limit_exceeded
      render json: { error: "slow_down", error_description: "Polling too frequently" }, status: :too_many_requests
    end
end
