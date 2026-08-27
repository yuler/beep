json.extract! beeper_install, :id, :title, :cron, :timezone, :status, :alert_state, :consecutive_failures, :config, :notification_channels, :ping_token, :last_ping_at, :next_run_at, :last_run_at, :created_at, :updated_at
if beeper_install.beeper
  json.beeper do
    json.extract! beeper_install.beeper, :id, :slug, :name, :version
  end
end
