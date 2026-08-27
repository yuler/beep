require "test_helper"

class BeeperTest < ActiveSupport::TestCase
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

  test "validates official beeper with valid manifest" do
    beeper = Beeper.new(
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
    assert beeper.valid?
    assert beeper.official?
  end

  test "validates custom beeper scoped to account" do
    account = accounts(:john_account)
    beeper = Beeper.new(
      account: account,
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert beeper.valid?
    assert beeper.custom?
    assert_not beeper.official?
  end

  test "enforces uniqueness of slug per account" do
    account = accounts(:john_account)
    Beeper.create!(
      account: account,
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )

    duplicate = Beeper.new(
      account: account,
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert_not duplicate.valid?
    assert_includes duplicate.errors[:slug], "has already been taken for this account"
  end

  test "allows same slug between official and custom beeper" do
    Beeper.create!(
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )

    custom = Beeper.new(
      account: accounts(:john_account),
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert custom.valid?
  end

  test "enforces slug matches manifest slug" do
    beeper = Beeper.new(
      slug: "different-slug",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    assert_not beeper.valid?
    assert_includes beeper.errors[:slug], "must match manifest slug 'custom-monitor'"
  end

  test "validates manifest schema contract" do
    invalid_manifest = @valid_manifest.merge("manifest_version" => 99, "schedule" => { "default_cron" => "invalid" })
    beeper = Beeper.new(
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: invalid_manifest
    )
    assert_not beeper.valid?
    assert beeper.errors[:manifest].present?
  end

  test "seed_official is idempotent" do
    assert_difference -> { Beeper.official.count }, 3 do
      Beeper.seed_official
    end

    site_uptime = Beeper.official.find_by(slug: "site-uptime")
    assert_not_nil site_uptime
    assert_equal "1.0.0", site_uptime.version
    assert_equal "*/5 * * * *", site_uptime.default_cron
    assert_equal 2, site_uptime.failure_threshold

    heartbeat = Beeper.official.find_by(slug: "heartbeat")
    assert_not_nil heartbeat
    assert heartbeat.webhook_ingest?

    # Running again should not create duplicate rows
    assert_no_difference -> { Beeper.official.count } do
      Beeper.seed_official
    end
  end

  test "produce_signal returns error when receiver class is missing" do
    beeper = Beeper.new(
      slug: "custom-monitor",
      version: "1.0.0",
      manifest: @valid_manifest
    )
    result = beeper.produce_signal(config: {})
    assert result.error?
    assert_equal "Beeper implementation not found", result.title
  end
end
