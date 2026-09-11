module Channel::Cli
  module_function

  def deliver_beep(channel, beep, run: nil)
    expires = (run&.scheduled_for || Time.current) + ChannelDelivery::DEFAULT_TTL
    delivery = channel.deliveries.create!(
      beep_run: run,
      payload: {
        id: run&.id,
        event: "beep.fired",
        source: beep.source_slug,
        source_type: beep.source_type,
        source_id: beep.source_id,
        intent: beep.intent,
        beep_id: beep.id,
        title: beep.title,
        body: beep.body_text,
        scheduled_for: (run&.scheduled_for || Time.current).iso8601,
        expires_at: expires.iso8601,
        metadata: beep.metadata || {}
      },
      expires_at: expires
    )
    { "channel_id" => channel.id, "channel_name" => channel.name, "delivery_id" => delivery.id, "status" => "queued" }
  rescue StandardError => error
    { "channel_id" => channel.id, "channel_name" => channel.name, "status" => "error", "error" => error.class.name }
  end

  def deliver_test!(channel)
    expires = Time.current + ChannelDelivery::DEFAULT_TTL
    channel.deliveries.create!(
      payload: {
        event: "beep.test",
        title: "Test notification",
        body: "This is a test notification for CLI channel #{channel.name}",
        expires_at: expires.iso8601
      },
      expires_at: expires
    )
  end

  def validate_config(channel)
  end
end
