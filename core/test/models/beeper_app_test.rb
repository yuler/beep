require "test_helper"

class BeeperAppTest < ActiveSupport::TestCase
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

  test "validates official beeper app with valid manifest" do
    beeper_app = BeeperApp.new(
      slug: "echo",
      version: "1.0.0",
      manifest: {
        "manifest_version" => 1,
        "slug" => "echo",
        "name" => "Echo",
        "version" => "1.0.0",
        "author" => "Beep",
        "schedule" => { "default_cron" => "*/5 * * * *" }
      }
    )
    assert beeper_app.valid?
    assert beeper_app.official?
  end

  test "validates custom beeper app scoped to account" do
    account = accounts(:john_account)
    beeper_app = BeeperApp.new(
      account: account,
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert beeper_app.valid?
    assert beeper_app.custom?
    assert_not beeper_app.official?
  end

  test "enforces uniqueness of slug per account" do
    account = accounts(:john_account)
    BeeperApp.create!(
      account: account,
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )

    duplicate = BeeperApp.new(
      account: account,
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert_not duplicate.valid?
    assert_includes duplicate.errors[:slug], "has already been taken for this account"
  end

  test "allows same slug between official and custom beeper app" do
    BeeperApp.create!(
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )

    custom = BeeperApp.new(
      account: accounts(:john_account),
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert custom.valid?
  end

  test "enforces slug matches manifest slug" do
    beeper_app = BeeperApp.new(
      slug: "different-slug",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert_not beeper_app.valid?
    assert_includes beeper_app.errors[:slug], "must match manifest slug 'custom-monitor'"
  end

  test "validates manifest schema contract" do
    invalid_manifest = @valid_manifest.merge("manifest_version" => 99, "schedule" => { "default_cron" => "invalid" })
    beeper_app = BeeperApp.new(
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: invalid_manifest
    )
    assert_not beeper_app.valid?
    assert beeper_app.errors[:manifest].present?
  end

  test "seed_official is idempotent" do
    assert_difference -> { BeeperApp.official.count }, 3 do
      BeeperApp.seed_official
    end

    site_uptime = BeeperApp.official.find_by(slug: "site-uptime")
    assert_not_nil site_uptime
    assert_equal "1.0.0", site_uptime.version
    assert_equal "*/5 * * * *", site_uptime.default_cron
    assert_equal 2, site_uptime.failure_threshold

    heartbeat = BeeperApp.official.find_by(slug: "heartbeat")
    assert_not_nil heartbeat
    assert heartbeat.webhook_ingest?

    # Running again should not create duplicate rows
    assert_no_difference -> { BeeperApp.official.count } do
      BeeperApp.seed_official
    end
  end

  test "produce_signal returns error when receiver class is missing" do
    beeper_app = BeeperApp.new(
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    result = beeper_app.produce_signal(config: {})
    assert result.error?
    assert_equal "Beeper app implementation not found", result.title
  end
end
