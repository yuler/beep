json.extract! beep, :id, :title, :body, :kind, :status, :cron, :run_at, :next_run_at, :last_run_at, :timezone, :alert_state, :consecutive_failures, :ping_token, :last_ping_at, :created_at
json.plugin_id beep.plugin_id
json.plugin_config beep.plugin_config
if beep.plugin
  json.plugin do
    json.extract! beep.plugin, :id, :slug, :name, :version, :description
  end
end
json.runs beep.runs.sort_by(&:scheduled_for).reverse do |run|
  json.partial! "api/v1/beeps/run", run: run
end
