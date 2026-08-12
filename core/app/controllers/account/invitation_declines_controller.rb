class Account::InvitationDeclinesController < ApplicationController
  allow_unauthorized_access
  before_action :set_invitation

  def create
    account_name = @account_invitation.account.name
    @account_invitation.decline!

    redirect_to my_accounts_url(script_name: nil),
      notice: "Declined invitation to #{account_name}."
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
