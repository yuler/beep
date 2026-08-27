class BeeperRun < ApplicationRecord
  self.table_name = "beeper_runs"

  SIGNAL_RESULT_MAX_BYTES = 8.kilobytes

  belongs_to :beeper_install

  enum :status, %w[ pending running succeeded failed skipped expired ].index_by(&:itself)
  enum :signal_status, %w[ ok alerting error ].index_by(&:itself)

  def deliver_later
    RunBeeperJob.perform_later(self)
  end

  def execute_now
    return unless claim_execution?

    if beeper_install.expired?(scheduled_for)
      update!(status: :expired)
      beeper_install.finish_firing(last_run_at: scheduled_for)
      return
    end

    signal = beeper_install.beeper.produce_signal(config: beeper_install.effective_config)
    sanitized_result = sanitize_signal_result(signal.to_h)

    decision = BeeperRun::AlertEvaluator.evaluate(
      install: beeper_install,
      signal: signal
    )

    update!(
      signal_status: signal.status.to_s,
      signal_result: sanitized_result
    )

    beeper_install.update!(
      alert_state: decision.next_alert_state,
      consecutive_failures: decision.next_consecutive_failures
    )

    if decision.should_notify
      beeper_install.notify_from!(signal)
    end

    update!(status: :succeeded)
    beeper_install.finish_firing(last_run_at: scheduled_for)
  rescue StandardError => e
    update!(
      signal_status: "error",
      signal_result: { "status" => "error", "message" => e.message },
      status: :failed
    )
    beeper_install.finish_firing(last_run_at: scheduled_for)
  end

  private

  def claim_execution?
    claimed = self.class.where(id: id, status: :pending).update_all(status: "running", updated_at: Time.current) == 1
    claimed || running?
  end

  def sanitize_signal_result(hash)
    json_str = hash.to_json
    if json_str.bytesize > SIGNAL_RESULT_MAX_BYTES
      {
        "status" => hash["status"],
        "title" => hash["title"],
        "message" => hash["message"]&.truncate(500),
        "metrics" => hash["metrics"],
        "truncated" => true
      }.compact
    else
      hash
    end
  end
end
