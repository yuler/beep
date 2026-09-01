json.runner do
  json.partial! "api/v1/runners/runner", runner: @runner
  json.token @runner.raw_token if @runner.raw_token.present?
end
