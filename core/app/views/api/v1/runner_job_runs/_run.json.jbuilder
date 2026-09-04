json.extract! run, :id, :runner_job_id, :runner_id, :scheduled_for, :status, :claimed_at, :result_status, :result, :created_at, :updated_at
json.log_preview run.log.to_s.last(500)
