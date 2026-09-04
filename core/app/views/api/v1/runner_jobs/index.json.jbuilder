json.jobs @jobs do |job|
  json.partial! "api/v1/runner_jobs/job", job: job
end
