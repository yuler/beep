json.extract! beeper, :id, :title, :body, :cron, :timezone, :status, :alert_state, :consecutive_failures, :config, :signal_metadata, :notification_channels, :ping_token, :last_ping_at, :next_run_at, :last_run_at, :created_at, :updated_at
if beeper.beeper_app
  json.beeper_app do
    json.extract! beeper.beeper_app, :id, :slug, :name, :version, :description, :inputs, :metrics
  end
end
json.runs beeper.runs.sort_by(&:scheduled_for).reverse do |run|
  json.partial! "api/v1/beepers/run", run: run
end
