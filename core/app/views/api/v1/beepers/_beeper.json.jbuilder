json.extract! beeper, :id, :slug, :version, :created_at
json.name beeper.name
json.description beeper.description
json.default_cron beeper.default_cron
json.failure_threshold beeper.failure_threshold
json.min_interval_seconds beeper.min_interval_seconds
json.webhook_ingest beeper.webhook_ingest?
json.inputs beeper.inputs
json.metrics beeper.metrics
