json.extract! beep, :id, :title, :body, :kind, :status, :run_at, :next_run_at, :last_run_at, :timezone, :created_at
json.runs beep.runs.sort_by(&:scheduled_for).reverse do |run|
  json.partial! "api/v1/beeps/run", run: run
end
