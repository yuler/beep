class BeeperRun::AlertEvaluator
  # Table from beeper-ecosystem.md:
  # | Previous alert_state | check_status       | Notify                            | Next state                                      |
  # | ok                   | ok                 | no                                | ok, counter 0                                   |
  # | ok                   | alerting / error   | only when counter + 1 >= threshold | alerting if notified, else ok (counter + 1)     |
  # | alerting             | alerting / error   | no (already alerting)             | alerting, counter + 1                           |
  # | alerting             | ok                 | yes — recovery                    | ok, counter 0                                   |

  Decision = Data.define(:should_notify, :next_alert_state, :next_consecutive_failures, :is_recovery)

  def self.evaluate(install:, check_result:)
    threshold = install.failure_threshold
    current_state = install.alert_state.to_s
    failures = install.consecutive_failures || 0

    if check_result.ok?
      if current_state == "alerting"
        # Recovery
        Decision.new(
          should_notify: true,
          next_alert_state: "ok",
          next_consecutive_failures: 0,
          is_recovery: true
        )
      else
        Decision.new(
          should_notify: false,
          next_alert_state: "ok",
          next_consecutive_failures: 0,
          is_recovery: false
        )
      end
    else
      # alerting or error
      new_failures = failures + 1
      if current_state == "alerting"
        Decision.new(
          should_notify: false,
          next_alert_state: "alerting",
          next_consecutive_failures: new_failures,
          is_recovery: false
        )
      else
        if new_failures >= threshold
          Decision.new(
            should_notify: true,
            next_alert_state: "alerting",
            next_consecutive_failures: new_failures,
            is_recovery: false
          )
        else
          Decision.new(
            should_notify: false,
            next_alert_state: "ok",
            next_consecutive_failures: new_failures,
            is_recovery: false
          )
        end
      end
    end
  end
end
