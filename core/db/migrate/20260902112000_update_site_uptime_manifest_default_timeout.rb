class UpdateSiteUptimeManifestDefaultTimeout < ActiveRecord::Migration[8.2]
  class MigrationBeeperApp < ApplicationRecord
    self.table_name = "beeper_apps"
  end

  def up
    app = MigrationBeeperApp.find_by(account_id: nil, slug: "site-uptime")
    return unless app&.manifest.is_a?(Hash)

    manifest = app.manifest
    timeout_input = manifest.dig("inputs")&.find { |i| i["name"] == "timeout_ms" }
    if timeout_input
      timeout_input["default"] = 5000
      app.update_columns(manifest: manifest)
    end
  end

  def down
    app = MigrationBeeperApp.find_by(account_id: nil, slug: "site-uptime")
    return unless app&.manifest.is_a?(Hash)

    manifest = app.manifest
    timeout_input = manifest.dig("inputs")&.find { |i| i["name"] == "timeout_ms" }
    if timeout_input
      timeout_input["default"] = 3000
      app.update_columns(manifest: manifest)
    end
  end
end
