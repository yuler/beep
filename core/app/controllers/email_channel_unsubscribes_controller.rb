class EmailChannelUnsubscribesController < ApplicationController
  allow_unauthenticated_access
  skip_before_action :require_account
  skip_forgery_protection only: :create

  before_action :set_user

  def show
  end

  def create
    if @user.notification_channel?("email")
      @user.update!(notification_channels: Array(@user.notification_channels) - %w[ email ])
    end
    @unsubscribed = true
    render :show
  end

  private
    def set_user
      @user = User.find_signed!(params[:token], purpose: :email_channel_unsubscribe)
    rescue ActiveSupport::MessageVerifier::InvalidSignature
      raise ActiveRecord::RecordNotFound
    end
end
