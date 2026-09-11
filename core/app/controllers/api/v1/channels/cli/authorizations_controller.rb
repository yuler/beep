class Api::V1::Channels::Cli::AuthorizationsController < Api::V1::BaseController
  disallow_account_scope
  allow_unauthenticated_access only: %i[ create ]

  def create
    @auth = ChannelAuthorization.create_request!(channel_name: params[:channel_name])
    @web_origin = Rails.configuration.x.web_origin
    render :create, status: :created
  end

  def show
    @auth = ChannelAuthorization.active.find_by(user_code: params[:user_code].to_s.upcase.strip)
    if @auth
      render :show, status: :ok
    else
      render_json_error(status: :not_found, message: "Invalid or expired user code", code: "NOT_FOUND")
    end
  end

  def update
    @auth = ChannelAuthorization.active.find_by(user_code: params[:user_code].to_s.upcase.strip)
    user = Current.user || Current.identity&.personal_user
    if @auth && user && @auth.approve!(user: user, name: params[:channel_name])
      @channel = @auth.channel
      render :update, status: :ok
    else
      render_json_error(status: :unprocessable_entity, message: "Unable to approve authorization", code: "UNPROCESSABLE_ENTITY")
    end
  end

  def destroy
    @auth = ChannelAuthorization.active.find_by(user_code: params[:user_code].to_s.upcase.strip)
    if @auth&.deny!
      head :no_content
    else
      render_json_error(status: :unprocessable_entity, message: "Unable to deny authorization", code: "UNPROCESSABLE_ENTITY")
    end
  end
end
