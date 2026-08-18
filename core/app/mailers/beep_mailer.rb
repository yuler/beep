class BeepMailer < ApplicationMailer
  def reminder(beep_run)
    @beep = beep_run.beep
    @account = @beep.account
    @unsubscribe_url = email_channel_unsubscribe_url(@account.email_channel_unsubscribe_token)

    headers["List-Unsubscribe"] = "<#{@unsubscribe_url}>"
    headers["List-Unsubscribe-Post"] = "List-Unsubscribe=One-Click"

    mail to: @account.owner_identity.email, subject: @beep.title
  end
end
