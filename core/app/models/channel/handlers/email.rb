class Channel::Handlers::Email < Channel::Handlers::Base
  class << self
    def deliver_beep(channel, beep, run: nil)
      user = channel.user
      if run
        BeepMailer.beep(run, user: user).deliver_now
      end
      { "channel_id" => channel.id, "status" => "sent" }
    rescue StandardError => error
      { "channel_id" => channel.id, "status" => "error", "error" => error.class.name }
    end

    def deliver_test!(channel)
      user = channel.user
      return unless user&.identity&.email.present?

      beep = channel.account.beeps.build(title: "Test notification", body: "This is a test notification for Email channel #{channel.name}")
      run = beep.runs.build(scheduled_for: Time.current)
      BeepMailer.beep(run, user: user).deliver_now
    end
  end
end
