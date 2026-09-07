json.status "ok"
json.pushed_count @pushed_jobs.size
json.jobs @pushed_jobs, partial: "api/v1/runner/jobs/job", as: :job
