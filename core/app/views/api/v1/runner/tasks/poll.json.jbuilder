json.task do
  json.id @run.id
  json.beeper_id @run.beeper_id
  json.title @run.beeper.title
  json.app_slug @run.beeper.beeper_app.slug
  json.manifest @run.beeper.beeper_app.manifest
  json.config @run.beeper.effective_config
  json.scheduled_for @run.scheduled_for.utc.iso8601
  json.timeout_seconds 30
end
