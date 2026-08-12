class Account::InvitationAcceptancesController < ApplicationController
  allow_unauthorized_access
  before_action :set_invitation

  def create
    @account_invitation.accept!

    write_last_account_slug(@account_invitation.account.slug)
    redirect_to root_url(script_name: @account_invitation.account.slug_path),
      notice: "You've joined #{@account_invitation.account.name}."
  rescue Account::Invitation::EmailMismatch => error
    redirect_to account_invitation_path(@account_invitation.token), alert: error.message
  rescue Account::Invitation::AlreadyResponded
    redirect_to account_invitation_path(@account_invitation.token),
      alert: "This invitation has already been responded to."
  end

  private
    def set_invitation
      @account_invitation = Current.account.invitations.find_by!(token: params.expect(:invitation_token))
    end
end
