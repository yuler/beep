class BeepMailer < ApplicationMailer
  def reminder(beep_run, user:)
    @beep = beep_run.beep
    @account = @beep.account
    @user = user
    @unsubscribe_url = email_channel_unsubscribe_url(user.email_channel_unsubscribe_token)

    headers["List-Unsubscribe"] = "<#{@unsubscribe_url}>"
    headers["List-Unsubscribe-Post"] = "List-Unsubscribe=One-Click"
    @run = beep_run
    subject = @run.check_result&.dig("title").presence || @beep.title
    mail to: user.identity.email, subject: subject
  end
end
