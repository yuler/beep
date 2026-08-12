class Account::InvitationsController < ApplicationController
  allow_unauthenticated_access only: :show
  allow_unauthorized_access only: :show
  before_action :ensure_admin, only: %i[ index new create ]
  before_action :set_invitation_by_token, only: :show

  def index
    @account_invitation = Account::Invitation.new
    @invitations = Current.account.invitations.includes(:acceptance, :decline, :invited_by)
  end

  def new
    @invitation = Current.account.invitations.new
  end

  def create
    @account_invitation = Current.account.invitations.new(**account_invitation_params)

    if @account_invitation.save
      redirect_to account_invitations_path, notice: "Invitation sent."
    else
      @invitation = @account_invitation
      @invitations = Current.account.invitations
      render :new, status: :unprocessable_entity
    end
  end

  def show
    if Current.identity.blank?
      session[:return_to_after_authenticating] = account_invitation_url(
        @account_invitation.token,
        script_name: @account_invitation.account.slug_path
      )
    end
  end

  private
    def set_invitation_by_token
      @account_invitation = Current.account.invitations.find_by!(token: params.expect(:token))
    end

    def account_invitation_params
      params.require(:account_invitation).permit(:email).merge(invited_by: Current.user, account: Current.account)
    end
end
