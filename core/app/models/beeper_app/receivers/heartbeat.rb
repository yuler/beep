class BeeperApp::Receivers::Heartbeat < BeeperApp::Receivers::Base
  DEFAULT_GRACE_PERIOD_MINUTES = 15

  def call
    grace_period_minutes = (config["grace_period_minutes"] || DEFAULT_GRACE_PERIOD_MINUTES).to_i
    last_ping_at = config["last_ping_at"].present? ? Time.zone.parse(config["last_ping_at"].to_s) : nil

    if last_ping_at.nil?
      beeper_created_at = config["beeper_created_at"].present? ? Time.zone.parse(config["beeper_created_at"].to_s) : nil
      if beeper_created_at && (Time.current - beeper_created_at) <= grace_period_minutes.minutes
        return BeeperApp::Signal.new(
          status: :ok,
          title: "Waiting for first heartbeat",
          message: "No ping received yet, within initial grace period of #{grace_period_minutes}m",
          metrics: { "minutes_since_last_ping" => nil }
        )
      else
        return BeeperApp::Signal.new(
          status: :alerting,
          title: "Heartbeat never received",
          message: "No ping has ever been received for this heartbeat monitor (grace period: #{grace_period_minutes}m)",
          metrics: { "minutes_since_last_ping" => nil }
        )
      end
    end

    seconds_since_ping = Time.current - last_ping_at
    minutes_since_ping = (seconds_since_ping / 60.0).round

    metrics = {
      "minutes_since_last_ping" => minutes_since_ping,
      "last_ping_at" => last_ping_at.utc.iso8601
    }

    if minutes_since_ping > grace_period_minutes
      BeeperApp::Signal.new(
        status: :alerting,
        title: "Heartbeat is missing",
        message: "Last ping was received #{minutes_since_ping} minutes ago (exceeded grace period of #{grace_period_minutes}m)",
        metrics: metrics
      )
    else
      BeeperApp::Signal.new(
        status: :ok,
        title: "Heartbeat is healthy",
        message: "Last ping was received #{minutes_since_ping} minutes ago (within grace period of #{grace_period_minutes}m)",
        metrics: metrics
      )
    end
  end
end
