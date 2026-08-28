class Beeper::AlertPolicy::Windowed < Beeper::AlertPolicy
  # Windowed policy: alerts if >= min_failures in the last window_size runs.
  DEFAULT_WINDOW_SIZE = 5
  DEFAULT_MIN_FAILURES = 3

  def evaluate(signal:)
    config = policy_config
    window_size = (config["window_size"] || DEFAULT_WINDOW_SIZE).to_i
    min_failures = (config["min_failures"] || DEFAULT_MIN_FAILURES).to_i

    current_state = beeper.alert_state.to_s
    failures = beeper.consecutive_failures || 0
    recoveries = beeper.consecutive_recoveries || 0

    # Fetch last (window_size - 1) completed runs to form window with current signal
    recent_runs = beeper.runs.where.not(status: [ :pending, :running ]).order(created_at: :desc).limit(window_size - 1)
    recent_signals = recent_runs.map(&:signal_status).compact

    all_signals_in_window = [ signal.status.to_s ] + recent_signals
    failure_count_in_window = all_signals_in_window.count { |s| s != "ok" }

    breached = failure_count_in_window >= min_failures

    if breached
      if current_state == "alerting"
        Decision.new(
          should_notify: false,
          next_alert_state: "alerting",
          next_consecutive_failures: failures + 1,
          next_consecutive_recoveries: 0,
          is_recovery: false
        )
      else
        Decision.new(
          should_notify: true,
          next_alert_state: "alerting",
          next_consecutive_failures: failures + 1,
          next_consecutive_recoveries: 0,
          is_recovery: false
        )
      end
    else
      if current_state == "alerting" || current_state == "recovering"
        # Window no longer breached -> recovery
        Decision.new(
          should_notify: true,
          next_alert_state: "ok",
          next_consecutive_failures: 0,
          next_consecutive_recoveries: 0,
          is_recovery: true
        )
      elsif failure_count_in_window > 0
        Decision.new(
          should_notify: false,
          next_alert_state: "pending",
          next_consecutive_failures: signal.ok? ? failures : failures + 1,
          next_consecutive_recoveries: 0,
          is_recovery: false
        )
      else
        Decision.new(
          should_notify: false,
          next_alert_state: "ok",
          next_consecutive_failures: 0,
          next_consecutive_recoveries: 0,
          is_recovery: false
        )
      end
    end
  end
end
