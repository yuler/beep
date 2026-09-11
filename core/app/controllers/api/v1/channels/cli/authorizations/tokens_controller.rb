class Api::V1::Channels::Cli::Authorizations::TokensController < Api::V1::BaseController
  disallow_account_scope
  allow_unauthenticated_access

  def create
    unless params[:grant_type] == "urn:ietf:params:oauth:grant-type:device_code"
      render json: { error: "unsupported_grant_type", error_description: "grant_type must be urn:ietf:params:oauth:grant-type:device_code" }, status: :bad_request
      return
    end

    @auth = ChannelAuthorization.find_by(device_code: params[:device_code].to_s.strip)
    unless @auth
      render json: { error: "invalid_grant", error_description: "Invalid device code" }, status: :bad_request
      return
    end

    @auth.poll!

    if @auth.status == "approved"
      @channel = @auth.channel
      render :create, status: :ok
    elsif @auth.status == "access_denied"
      render json: { error: "access_denied", error_description: "The user denied the authorization request" }, status: :bad_request
    elsif @auth.expired? || @auth.status == "expired"
      render json: { error: "expired_token", error_description: "The device code has expired" }, status: :bad_request
    else
      render json: { error: "authorization_pending", error_description: "The authorization request is still pending" }, status: :bad_request
    end
  end
end
