class BeepMailer < ApplicationMailer
  def reminder(beep_run, user:)
    @beep = beep_run.beep
    @account = @beep.account
    @user = user
    @unsubscribe_url = email_channel_unsubscribe_url(user.email_channel_unsubscribe_token)

    headers["List-Unsubscribe"] = "<#{@unsubscribe_url}>"
    headers["List-Unsubscribe-Post"] = "List-Unsubscribe=One-Click"

    mail to: user.identity.email, subject: @beep.title
  end
end
