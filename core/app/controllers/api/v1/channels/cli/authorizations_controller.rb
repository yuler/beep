class Api::V1::Channels::Cli::AuthorizationsController < Api::V1::BaseController
  skip_account_scope only: %i[ create show destroy ]
  allow_unauthenticated_access only: %i[ create ]
  rate_limit to: 20, within: 1.minute, only: %i[ create ],
    by: -> { request.remote_ip }, with: :rate_limit_exceeded
  rate_limit to: 30, within: 1.minute, only: %i[ show update destroy ],
    by: -> { params[:user_code].to_s.upcase.strip.presence || request.remote_ip },
    with: :rate_limit_exceeded

  def create
    @auth = Channel::Authorization.create_request!(channel_name: params[:channel_name])
    @web_origin = Rails.configuration.x.web_origin
    render :create, status: :created
  end

  def show
    @auth = Channel::Authorization.active.find_by(user_code: params[:user_code].to_s.upcase.strip)
    if @auth
      render :show, status: :ok
    else
      render_json_error(status: :not_found, message: "Invalid or expired user code", code: "NOT_FOUND")
    end
  end

  def update
    @auth = Channel::Authorization.active.find_by(user_code: params[:user_code].to_s.upcase.strip)
    user = Current.user
    if @auth && user && @auth.approve!(user: user, name: params[:channel_name])
      @channel = @auth.channel
      render :update, status: :ok
    else
      render_json_error(status: :unprocessable_entity, message: "Unable to approve authorization", code: "UNPROCESSABLE_ENTITY")
    end
  end

  def destroy
    @auth = Channel::Authorization.active.find_by(user_code: params[:user_code].to_s.upcase.strip)
    if @auth&.deny!
      head :no_content
    else
      render_json_error(status: :unprocessable_entity, message: "Unable to deny authorization", code: "UNPROCESSABLE_ENTITY")
    end
  end

  private
    def rate_limit_exceeded
      render json: { error: "slow_down", error_description: "Too many requests" }, status: :too_many_requests
    end
end
