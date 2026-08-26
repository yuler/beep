require "test_helper"

class PluginTest < ActiveSupport::TestCase
  setup do
    @valid_manifest = {
      "manifest_version" => 1,
      "slug" => "custom-monitor",
      "name" => "Custom Monitor",
      "version" => "1.0.0",
      "description" => "A test monitor",
      "author" => "Community",
      "schedule" => {
        "default_cron" => "*/10 * * * *",
        "min_interval_seconds" => 60,
        "failure_threshold" => 3
      },
      "ingest" => {
        "webhook" => false
      },
      "inputs" => [
        {
          "name" => "url",
          "label" => "Target URL",
          "type" => "url",
          "required" => true
        }
      ],
      "metrics" => [
        {
          "name" => "status",
          "label" => "Status Code",
          "type" => "number"
        }
      ]
    }
  end

  test "validates official plugin with valid manifest" do
    plugin = Plugin.new(
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert plugin.valid?
    assert plugin.official?
    assert_not plugin.custom?
  end

  test "validates custom plugin scoped to account" do
    account = accounts(:john_account)
    plugin = Plugin.new(
      account: account,
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert plugin.valid?
    assert plugin.custom?
    assert_not plugin.official?
  end

  test "enforces uniqueness of slug per account" do
    account = accounts(:john_account)
    Plugin.create!(
      account: account,
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )

    duplicate = Plugin.new(
      account: account,
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert_not duplicate.valid?
    assert_includes duplicate.errors[:slug], "has already been taken for this account"
  end

  test "allows same slug between official and custom plugin" do
    Plugin.create!(
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )

    custom = Plugin.new(
      account: accounts(:john_account),
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert custom.valid?
  end

  test "enforces slug matches manifest slug" do
    plugin = Plugin.new(
      slug: "different-slug",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert_not plugin.valid?
    assert_includes plugin.errors[:slug], "must match manifest slug 'custom-monitor'"
  end

  test "validates manifest schema contract" do
    invalid_manifest = @valid_manifest.merge("manifest_version" => 99, "schedule" => { "default_cron" => "invalid" })
    plugin = Plugin.new(
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: invalid_manifest
    )
    assert_not plugin.valid?
    assert plugin.errors[:manifest].present?
  end

  test "seed_official_plugins! is idempotent" do
    assert_difference -> { Plugin.official.count }, 3 do
      Plugin.seed_official_plugins!
    end

    site_uptime = Plugin.official.find_by(slug: "site-uptime")
    assert_not_nil site_uptime
    assert_equal "1.0.0", site_uptime.version
    assert_equal "*/5 * * * *", site_uptime.default_cron
    assert_equal 2, site_uptime.failure_threshold

    heartbeat = Plugin.official.find_by(slug: "heartbeat")
    assert_not_nil heartbeat
    assert heartbeat.webhook_ingest?

    # Running again should not create duplicate rows
    assert_no_difference -> { Plugin.official.count } do
      Plugin.seed_official_plugins!
    end
  end
end
