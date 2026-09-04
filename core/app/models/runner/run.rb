class Runner::Run < ApplicationRecord
  RESULT_MAX_BYTES = 8.kilobytes
  LOG_MAX_BYTES = 256.kilobytes

  belongs_to :runner_job, class_name: "Runner::Job", foreign_key: :runner_job_id, inverse_of: :runs
  belongs_to :runner

  enum :status, %w[ pending running succeeded failed expired ].index_by(&:itself)
  enum :result_status, %w[ ok alerting error ].index_by(&:itself)

  def claim_for!(runner)
    claimed = self.class.where(id: id, status: :pending).update_all(
      status: "running",
      runner_id: runner.id,
      claimed_at: Time.current,
      updated_at: Time.current
    ) == 1

    if claimed
      reload
      true
    else
      false
    end
  end

  def append_log!(chunk)
    text = chunk.to_s
    return if text.blank?

    with_lock do
      combined = "#{log}#{text}"
      if combined.bytesize > LOG_MAX_BYTES
        overflow = combined.bytesize - LOG_MAX_BYTES
        combined = combined.byteslice(overflow, LOG_MAX_BYTES).scrub("")
      end
      update!(log: combined)
    end
  end

  def record_result!(status:, title: nil, message: nil, metrics: {}, run_status: :succeeded)
    return false unless pending? || running?

    result_status = status.to_s.downcase
    result_status = "error" unless result_status.in?(%w[ ok alerting error ])
    run_status = :failed if result_status == "error" && run_status == :succeeded

    payload = sanitize_result(
      "status" => result_status,
      "title" => title.to_s.presence,
      "message" => message.to_s.presence,
      "metrics" => metrics.is_a?(Hash) ? metrics : {}
    )

    update!(
      result_status: result_status,
      result: payload,
      status: run_status
    )
    runner_job.finish_firing(last_run_at: scheduled_for)
    true
  end

  private

  def sanitize_result(hash)
    json_str = hash.to_json
    return hash if json_str.bytesize <= RESULT_MAX_BYTES

    truncated = {
      "status" => hash["status"],
      "title" => hash["title"]&.to_s&.truncate(200),
      "message" => hash["message"]&.to_s&.truncate(500),
      "metrics" => hash["metrics"].is_a?(Hash) ? hash["metrics"].slice(*hash["metrics"].keys.first(20)) : {},
      "truncated" => true
    }.compact

    truncated.delete("metrics") if truncated.to_json.bytesize > RESULT_MAX_BYTES
    truncated
  end
end
