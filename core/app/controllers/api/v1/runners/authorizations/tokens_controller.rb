class Api::V1::Runners::Authorizations::TokensController < Api::V1::BaseController
  disallow_account_scope
  allow_unauthenticated_access
  rate_limit to: 30, within: 1.minute, only: :create,
    by: -> { request.remote_ip },
    with: :rate_limit_exceeded

  def create
    unless params[:grant_type] == "urn:ietf:params:oauth:grant-type:device_code"
      return device_flow_error("unsupported_grant_type", "grant_type must be urn:ietf:params:oauth:grant-type:device_code")
    end

    @auth = Runner::Authorization.find_by(device_code: params[:device_code].to_s.strip)
    unless @auth
      return device_flow_error("invalid_grant", "Invalid device code")
    end

    if @auth.poll_interval_exceeded?
      return device_flow_error("slow_down", "Polling too frequently")
    end

    @auth.poll!

    if @auth.expired? || @auth.status == "expired"
      device_flow_error("expired_token", "The device code has expired")
    elsif @auth.status == "approved"
      @runner = @auth.consume_token!
      if @runner
        render :create, status: :ok
      else
        device_flow_error("invalid_grant", "Invalid device code")
      end
    elsif @auth.status == "consumed"
      device_flow_error("invalid_grant", "Invalid device code")
    elsif @auth.status == "access_denied"
      device_flow_error("access_denied", "The user denied the authorization request")
    else
      device_flow_error("authorization_pending", "The authorization request is still pending")
    end
  end

  private
    def device_flow_error(error, description, status: :bad_request)
      @device_flow_error = error
      @device_flow_error_description = description
      render "api/v1/runners/authorizations/device_flow_error", status: status
    end

    def rate_limit_exceeded
      device_flow_error("slow_down", "Polling too frequently", status: :too_many_requests)
    end
end
