class Beeper::Checkers::Heartbeat < Beeper::Checkers::Base
  DEFAULT_GRACE_PERIOD_MINUTES = 15

  def call
    grace_period_minutes = (config["grace_period_minutes"] || DEFAULT_GRACE_PERIOD_MINUTES).to_i
    last_ping_at = config["last_ping_at"].present? ? Time.zone.parse(config["last_ping_at"].to_s) : nil

    if last_ping_at.nil?
      return Beeper::CheckResult.new(
        status: :alerting,
        title: "Heartbeat never received",
        message: "No ping has ever been received for this heartbeat monitor (grace period: #{grace_period_minutes}m)",
        metrics: { "minutes_since_last_ping" => nil }
      )
    end

    seconds_since_ping = Time.current - last_ping_at
    minutes_since_ping = (seconds_since_ping / 60.0).round

    metrics = {
      "minutes_since_last_ping" => minutes_since_ping,
      "last_ping_at" => last_ping_at.utc.iso8601
    }

    if minutes_since_ping > grace_period_minutes
      Beeper::CheckResult.new(
        status: :alerting,
        title: "Heartbeat is missing",
        message: "Last ping was received #{minutes_since_ping} minutes ago (exceeded grace period of #{grace_period_minutes}m)",
        metrics: metrics
      )
    else
      Beeper::CheckResult.new(
        status: :ok,
        title: "Heartbeat is healthy",
        message: "Last ping was received #{minutes_since_ping} minutes ago (within grace period of #{grace_period_minutes}m)",
        metrics: metrics
      )
    end
  end
end
