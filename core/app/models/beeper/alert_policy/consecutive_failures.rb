class Beeper::AlertPolicy::ConsecutiveFailures < Beeper::AlertPolicy
  # 4-State Machine:
  # | Current state | Signal status | Next state | Notify? |
  # | ok            | ok            | ok         | No      |
  # | ok            | alert/error   | pending (if failures < threshold) or alerting | Yes if alerting |
  # | pending       | ok            | ok         | No      |
  # | pending       | alert/error   | pending (if failures < threshold) or alerting | Yes if threshold reached |
  # | alerting      | alert/error   | alerting   | No      |
  # | alerting      | ok            | recovering (if recovery_threshold > 1) or ok | Yes if ok (recovery) |
  # | recovering    | alert/error   | alerting   | No      |
  # | recovering    | ok            | recovering (if recoveries < threshold) or ok | Yes if ok (recovery) |

  def evaluate(signal:)
    threshold = beeper.failure_threshold
    rec_threshold = beeper.recovery_threshold
    current_state = beeper.alert_state.to_s
    failures = beeper.consecutive_failures || 0
    recoveries = beeper.consecutive_recoveries || 0

    if signal.ok?
      if current_state == "alerting" || current_state == "recovering"
        new_recoveries = (current_state == "alerting" ? 0 : recoveries) + 1
        if new_recoveries >= rec_threshold
          # Fully recovered
          Decision.new(
            should_notify: true,
            next_alert_state: "ok",
            next_consecutive_failures: 0,
            next_consecutive_recoveries: 0,
            is_recovery: true
          )
        else
          # Still recovering
          Decision.new(
            should_notify: false,
            next_alert_state: "recovering",
            next_consecutive_failures: failures,
            next_consecutive_recoveries: new_recoveries,
            is_recovery: false
          )
        end
      else
        # Was ok or pending -> reset to healthy ok
        Decision.new(
          should_notify: false,
          next_alert_state: "ok",
          next_consecutive_failures: 0,
          next_consecutive_recoveries: 0,
          is_recovery: false
        )
      end
    else
      # Signal is alerting or error
      new_failures = (current_state == "recovering" ? failures : failures) + 1

      if current_state == "alerting"
        Decision.new(
          should_notify: false,
          next_alert_state: "alerting",
          next_consecutive_failures: new_failures,
          next_consecutive_recoveries: 0,
          is_recovery: false
        )
      elsif current_state == "recovering"
        # Failed while recovering -> back to alerting
        Decision.new(
          should_notify: false,
          next_alert_state: "alerting",
          next_consecutive_failures: new_failures,
          next_consecutive_recoveries: 0,
          is_recovery: false
        )
      else
        # Was ok or pending
        if new_failures >= threshold
          Decision.new(
            should_notify: true,
            next_alert_state: "alerting",
            next_consecutive_failures: new_failures,
            next_consecutive_recoveries: 0,
            is_recovery: false
          )
        else
          Decision.new(
            should_notify: false,
            next_alert_state: "pending",
            next_consecutive_failures: new_failures,
            next_consecutive_recoveries: 0,
            is_recovery: false
          )
        end
      end
    end
  end
end
