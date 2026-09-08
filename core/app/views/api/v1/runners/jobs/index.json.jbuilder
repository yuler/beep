json.jobs @jobs do |job|
  json.partial! "api/v1/runners/jobs/job", job: job
end
