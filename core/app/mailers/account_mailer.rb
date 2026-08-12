class AccountMailer < ApplicationMailer
  def invite(invitation)
    @invitation = invitation
    @account = invitation.account
    @invited_by = invitation.invited_by

    mail subject: "#{@invited_by.name} invited you to join #{@account.name}", to: invitation.email
  end
end
