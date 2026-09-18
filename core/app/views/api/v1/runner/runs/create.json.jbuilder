json.run do
  json.id @run.id
  json.job_id @run.runner_job_id
  json.job_slug @run.runner_job.slug
  json.name @run.runner_job.name
  json.config @run.runner_job.config
  json.scheduled_for @run.scheduled_for.utc.iso8601
  json.timeout_seconds @run.runner_job.timeout_seconds
  json.log_url "#{@api_base_url}/api/v1/runner/runs/#{@run.id}/logs"
  json.result_url "#{@api_base_url}/api/v1/runner/runs/#{@run.id}/result"
end
