class SeedOfficialBeeperApps < ActiveRecord::Migration[8.2]
  class MigrationBeeperApp < ApplicationRecord
    self.table_name = "beeper_apps"
  end

  OFFICIAL_MANIFESTS = [
    {
      "manifest_version" => 1,
      "slug" => "site-uptime",
      "name" => "Site Uptime & Health Check",
      "version" => "1.0.0",
      "description" => "HTTP status and latency probe",
      "author" => "Beep Official",
      "schedule" => {
        "default_cron" => "*/5 * * * *",
        "min_interval_seconds" => 60
      },
      "alerting" => {
        "policy" => "consecutive_failures",
        "failure_threshold" => 2,
        "recovery_threshold" => 1
      },
      "inputs" => [
        {
          "name" => "target_url",
          "label" => "Target URL",
          "type" => "url",
          "required" => true,
          "placeholder" => "https://example.com"
        },
        {
          "name" => "expected_status",
          "label" => "Expected status",
          "type" => "number",
          "default" => 200,
          "min" => 100,
          "max" => 599
        },
        {
          "name" => "timeout_ms",
          "label" => "Timeout (ms)",
          "type" => "number",
          "default" => 5000,
          "max" => 10_000
        }
      ],
      "metrics" => [
        {
          "name" => "latency_ms",
          "label" => "Latency",
          "type" => "number",
          "unit" => "ms"
        },
        {
          "name" => "status",
          "label" => "HTTP status",
          "type" => "number"
        }
      ]
    },
    {
      "manifest_version" => 1,
      "slug" => "ssl-expiry",
      "name" => "SSL / TLS Certificate Expiry",
      "version" => "1.0.0",
      "description" => "Monitors SSL/TLS certificates and alerts before expiration",
      "author" => "Beep Official",
      "schedule" => {
        "default_cron" => "0 8 * * *",
        "min_interval_seconds" => 3600
      },
      "alerting" => {
        "policy" => "consecutive_failures",
        "failure_threshold" => 2,
        "recovery_threshold" => 1
      },
      "inputs" => [
        {
          "name" => "hostname",
          "label" => "Domain / Hostname",
          "type" => "string",
          "required" => true,
          "placeholder" => "example.com"
        },
        {
          "name" => "port",
          "label" => "Port",
          "type" => "number",
          "default" => 443,
          "min" => 1,
          "max" => 65_535
        },
        {
          "name" => "alert_days_before",
          "label" => "Alert when days remaining less than",
          "type" => "number",
          "default" => 14,
          "min" => 1,
          "max" => 90
        }
      ],
      "metrics" => [
        {
          "name" => "days_remaining",
          "label" => "Days Remaining",
          "type" => "number",
          "unit" => "days"
        }
      ]
    },
    {
      "manifest_version" => 1,
      "slug" => "heartbeat",
      "name" => "Dead Man's Snitch / Heartbeat",
      "version" => "1.0.0",
      "description" => "Monitors periodic cron jobs and pings, alerts when silent",
      "author" => "Beep Official",
      "schedule" => {
        "default_cron" => "*/15 * * * *",
        "min_interval_seconds" => 60
      },
      "alerting" => {
        "policy" => "consecutive_failures",
        "failure_threshold" => 1,
        "recovery_threshold" => 1
      },
      "capabilities" => [
        "webhook_ping"
      ],
      "inputs" => [
        {
          "name" => "grace_period_minutes",
          "label" => "Grace Period (minutes)",
          "type" => "number",
          "default" => 15,
          "min" => 1,
          "max" => 1440
        }
      ],
      "metrics" => [
        {
          "name" => "minutes_since_last_ping",
          "label" => "Minutes Since Last Ping",
          "type" => "number",
          "unit" => "min"
        }
      ]
    }
  ].freeze

  def up
    say_with_time "Seeding official beeper apps" do
      OFFICIAL_MANIFESTS.each do |manifest|
        beeper_app = MigrationBeeperApp.find_or_initialize_by(slug: manifest["slug"], account_id: nil)
        beeper_app.version = manifest["version"]
        beeper_app.manifest = manifest
        beeper_app.save!
      end
    end
  end

  def down
    slugs = OFFICIAL_MANIFESTS.map { |manifest| manifest["slug"] }
    MigrationBeeperApp.where(account_id: nil, slug: slugs).delete_all
  end
end
