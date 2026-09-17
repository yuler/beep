json.extract! beep, :id, :title, :body, :kind, :status, :cron, :run_at, :next_run_at, :last_run_at, :timezone, :notification_channels, :beeper_id, :source_type, :source_id, :intent, :metadata, :created_at
json.source beep.source_slug
if beep.beeper&.beeper_app
  json.beeper do
    json.extract! beep.beeper.beeper_app, :slug, :name
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
    json.partial! "api/v1/beeps/run", run: run
  end
else
  json.runs beep.runs.order(scheduled_for: :desc).limit(5) do |run|
    json.partial! "api/v1/beeps/run", run: run
  end
end
