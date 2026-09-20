class BeepMailer < ApplicationMailer
  def beep(beep_run, user:)
    @beep = beep_run.beep
    @account = @beep.account
    @user = user
    @unsubscribe_url = email_channel_unsubscribe_url(user.email_channel_unsubscribe_token)

    headers["List-Unsubscribe"] = "<#{@unsubscribe_url}>"
    headers["List-Unsubscribe-Post"] = "List-Unsubscribe=One-Click"
    @run = beep_run

    timezone = user.timezone.presence || @beep.timezone.presence || "UTC"
    zone = Time.find_zone(timezone) || Time.zone
    time = (@run.scheduled_for || Time.current).in_time_zone(zone)
    @formatted_time = time.strftime("%m-%d %H:%M")
    @detailed_time = time.strftime("%Y-%m-%d %H:%M %Z")

    mail to: user.identity.email, subject: "[Beep] #{@beep.title} (#{@formatted_time})"
  end
end
