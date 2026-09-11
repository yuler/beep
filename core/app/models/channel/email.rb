module Channel::Email
  module_function

  def deliver_beep(channel, beep, run: nil)
    user = channel.user
    if run
      BeepMailer.reminder(run, user: user).deliver_now
    end
    { "channel_id" => channel.id, "status" => "sent" }
  rescue StandardError => error
    { "channel_id" => channel.id, "status" => "error", "error" => error.class.name }
  end

  def deliver_test!(channel)
  end

  def validate_config(channel)
  end
end
