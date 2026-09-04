json.status "ok"
json.synced_count @synced_jobs.size
json.jobs @synced_jobs, partial: "api/v1/runner/jobs/job", as: :job
