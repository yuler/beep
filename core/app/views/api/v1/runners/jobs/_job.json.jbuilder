json.extract! job, :id, :runner_id, :name, :slug, :cron, :timezone, :status, :timeout_seconds, :config, :next_run_at, :last_run_at, :created_at, :updated_at
json.runner_online job.runner.online?
