json.extract! plugin, :id, :slug, :version, :created_at
json.name plugin.name
json.description plugin.description
json.default_cron plugin.default_cron
json.failure_threshold plugin.failure_threshold
json.min_interval_seconds plugin.min_interval_seconds
json.webhook_ingest plugin.webhook_ingest?
json.inputs plugin.inputs
json.metrics plugin.metrics
