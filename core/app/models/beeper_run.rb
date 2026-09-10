class BeeperRun < ApplicationRecord
  self.table_name = "beeper_runs"

  SIGNAL_RESULT_MAX_BYTES = 8.kilobytes
  RECENT_LIMIT = 5
  LIST_LIMIT = 50
  RETENTION = 30.days

  belongs_to :beeper

  enum :status, %w[ pending running succeeded failed skipped expired ].index_by(&:itself)
  enum :signal_status, %w[ ok alerting error ].index_by(&:itself)

  # { beeper_id => { total:, succeeded: } } for the given beepers, in one grouped query.
  def self.stats_by_beeper(beeper_ids)
    where(beeper_id: beeper_ids)
      .group(:beeper_id)
      .pluck(:beeper_id, Arel.sql("COUNT(*)"), Arel.sql("SUM(CASE WHEN status = 'succeeded' THEN 1 ELSE 0 END)"))
      .to_h { |beeper_id, total, succeeded| [ beeper_id, { total: total, succeeded: succeeded || 0 } ] }
  end

  # { beeper_id => [newest-first runs] } capped at RECENT_LIMIT runs per beeper, in one windowed query.
  def self.recent_by_beeper(beeper_ids, limit: RECENT_LIMIT)
    ranked = where(beeper_id: beeper_ids)
      .select("beeper_runs.*, ROW_NUMBER() OVER (PARTITION BY beeper_id ORDER BY scheduled_for DESC) AS run_rank")
    from(ranked, :beeper_runs)
      .where(run_rank: 1..limit)
      .group_by(&:beeper_id)
  end

  def self.prune_expired_now
    where(scheduled_for: ..RETENTION.ago).in_batches(of: 1_000) do |batch|
      batch.delete_all
    end
  end

  def execute_later
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
    record_signal_result!(signal)
  rescue StandardError => e
    update!(
      signal_status: "error",
      signal_result: { "status" => "error", "message" => e.message },
      status: :failed
    )
    beeper.finish_firing(last_run_at: scheduled_for)
  end

  def record_signal_result!(signal, run_status: :succeeded)
    sanitized_result = sanitize_signal_result(signal.to_h)
    decision = Beeper::AlertPolicy.for(beeper).evaluate(signal: signal)

    ApplicationRecord.transaction do
      update!(
        signal_status: signal.status.to_s,
        signal_result: sanitized_result,
        status: run_status
      )

      beeper.update!(
        alert_state: decision.next_alert_state,
        consecutive_failures: decision.next_consecutive_failures,
        consecutive_recoveries: decision.next_consecutive_recoveries
      )

      beeper.notify_from!(signal) if decision.should_notify
    end

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
