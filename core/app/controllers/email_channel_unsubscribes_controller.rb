class EmailChannelUnsubscribesController < ApplicationController
  allow_unauthenticated_access
  skip_before_action :require_account
  skip_forgery_protection only: :create

  before_action :set_account

  def show
  end

  def create
    @account.update!(email_channel_enabled: false) if @account.email_channel_enabled?
    @unsubscribed = true
    render :show
  end

  private
    def set_account
      @account = Account.find_signed!(params[:token], purpose: :email_channel_unsubscribe)
    rescue ActiveSupport::MessageVerifier::InvalidSignature
      raise ActiveRecord::RecordNotFound
    end
end
