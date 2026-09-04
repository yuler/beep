json.task do
  json.id @run.id
  json.job_id @run.runner_job_id
  json.job_slug @run.runner_job.slug
  json.name @run.runner_job.name
  json.config @run.runner_job.config
  json.scheduled_for @run.scheduled_for.utc.iso8601
  json.timeout_seconds @run.runner_job.timeout_seconds
  json.log_url "#{request.base_url}/api/v1/runner/tasks/#{@run.id}/logs"
  json.result_url "#{request.base_url}/api/v1/runner/tasks/#{@run.id}/result"
end
