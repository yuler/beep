json.extract! beeper_app, :id, :slug, :version, :created_at
json.name beeper_app.name
json.description beeper_app.description
json.default_cron beeper_app.default_cron
json.failure_threshold beeper_app.failure_threshold
json.min_interval_seconds beeper_app.min_interval_seconds
json.webhook_ingest beeper_app.webhook_ingest?
json.inputs beeper_app.inputs
json.metrics beeper_app.metrics
