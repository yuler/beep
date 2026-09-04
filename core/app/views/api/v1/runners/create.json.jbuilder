json.runner do
  json.partial! "api/v1/runners/runner", runner: @runner
  json.token @runner.token if @runner.token.present?
end
