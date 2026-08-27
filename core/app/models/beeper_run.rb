class BeeperRun < ApplicationRecord
  self.table_name = "beeper_runs"

  CHECK_RESULT_MAX_BYTES = 8.kilobytes

  belongs_to :beeper_install

  enum :status, %w[ pending running succeeded failed skipped expired ].index_by(&:itself)
  enum :check_status, %w[ ok alerting error ].index_by(&:itself)

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

    check_result = beeper_install.beeper.run_check(config: beeper_install.effective_config)
    sanitized_result = sanitize_check_result(check_result.to_h)

    decision = BeeperRun::AlertEvaluator.evaluate(
      install: beeper_install,
      check_result: check_result
    )

    update!(
      check_status: check_result.status.to_s,
      check_result: sanitized_result
    )

    beeper_install.update!(
      alert_state: decision.next_alert_state,
      consecutive_failures: decision.next_consecutive_failures
    )

    if decision.should_notify
      beeper_install.notify_from!(check_result)
    end

    update!(status: :succeeded)
    beeper_install.finish_firing(last_run_at: scheduled_for)
  rescue StandardError => e
    update!(
      check_status: "error",
      check_result: { "status" => "error", "message" => e.message },
      status: :failed
    )
    beeper_install.finish_firing(last_run_at: scheduled_for)
  end

  private

  def claim_execution?
    claimed = self.class.where(id: id, status: :pending).update_all(status: "running", updated_at: Time.current) == 1
    claimed || running?
  end

  def sanitize_check_result(hash)
    json_str = hash.to_json
    if json_str.bytesize > CHECK_RESULT_MAX_BYTES
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
