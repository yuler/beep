json.extract! beep, :id, :title, :body, :kind, :status, :cron, :run_at, :next_run_at, :last_run_at, :timezone, :notification_channels, :beeper_id, :source_type, :source_id, :intent, :metadata, :created_at
json.source beep.source_slug
if beep.beeper&.beeper_app
  json.beeper do
    json.extract! beep.beeper.beeper_app, :slug, :name
  end
end
json.runs beep.runs.sort_by(&:scheduled_for).reverse do |run|
  json.partial! "api/v1/beeps/run", run: run
end
