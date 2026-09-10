json.extract! beeper, :id, :title, :body, :cron, :timezone, :status, :alert_state, :consecutive_failures, :config, :signal_metadata, :notification_channels, :ping_token, :last_ping_at, :next_run_at, :last_run_at, :created_at, :updated_at
if beeper.beeper_app
  json.beeper_app do
    json.extract! beeper.beeper_app, :id, :slug, :name, :version, :description, :inputs, :metrics
  end
end
if local_assigns[:run_stats]
  json.run_stats do
    json.total run_stats[:total]
    json.succeeded run_stats[:succeeded]
  end
end
if local_assigns.key?(:runs)
  json.runs runs do |run|
    json.partial! "api/v1/beepers/run", run: run
  end
end
