class BeeperRun < ApplicationRecord
  self.table_name = "beeper_runs"

  SIGNAL_RESULT_MAX_BYTES = 8.kilobytes

  belongs_to :beeper

  enum :status, %w[ pending running succeeded failed skipped expired ].index_by(&:itself)
  enum :signal_status, %w[ ok alerting error ].index_by(&:itself)

  def deliver_later
    RunBeeperJob.perform_later(self)
  end

  def execute_now
    return unless claim_execution?

    if beeper.expired?(scheduled_for)
      update!(status: :expired)
      beeper.finish_firing(last_run_at: scheduled_for)
      return
    end

    signal = beeper.beeper_app.produce_signal(config: beeper.effective_config)
    sanitized_result = sanitize_signal_result(signal.to_h)

    decision = Beeper::AlertPolicy.for(beeper).evaluate(signal: signal)

    update!(
      signal_status: signal.status.to_s,
      signal_result: sanitized_result
    )

    beeper.update!(
      alert_state: decision.next_alert_state,
      consecutive_failures: decision.next_consecutive_failures,
      consecutive_recoveries: decision.next_consecutive_recoveries
    )

    if decision.should_notify
      beeper.notify_from!(signal)
    end

    update!(status: :succeeded)
    beeper.finish_firing(last_run_at: scheduled_for)
  rescue StandardError => e
    update!(
      signal_status: "error",
      signal_result: { "status" => "error", "message" => e.message },
      status: :failed
    )
    beeper.finish_firing(last_run_at: scheduled_for)
  end

  private

  def claim_execution?
    claimed = self.class.where(id: id, status: :pending).update_all(status: "running", updated_at: Time.current) == 1
    claimed || running?
  end

  def sanitize_signal_result(hash)
    json_str = hash.to_json
    return hash if json_str.bytesize <= SIGNAL_RESULT_MAX_BYTES

    sanitized_metrics = sanitize_metrics(hash["metrics"])
    truncated_hash = {
      "status" => hash["status"],
      "title" => hash["title"]&.to_s&.truncate(200),
      "message" => hash["message"]&.to_s&.truncate(500),
      "metrics" => sanitized_metrics,
      "truncated" => true
    }.compact

    if truncated_hash.to_json.bytesize > SIGNAL_RESULT_MAX_BYTES
      truncated_hash.delete("metrics")
    end

    truncated_hash
  end

  def sanitize_metrics(metrics)
    return nil unless metrics.is_a?(Hash)

    # Keep at most 20 scalar metrics entries, truncate long string values
    metrics.slice(*metrics.keys.first(20)).transform_values do |val|
      if val.is_a?(String)
        val.truncate(100)
      elsif val.is_a?(Numeric) || val.is_a?(TrueClass) || val.is_a?(FalseClass)
        val
      else
        val.to_s.truncate(100)
      end
    end
  end
end
