require "test_helper"

class Api::V1::BeeperApps::Heartbeat::PingsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @account = accounts(:john_account)
    @beeper_app = BeeperApp.create!(
      slug: "heartbeat",
      version: "1.0.0",
      manifest: {
        "manifest_version" => 1,
        "slug" => "heartbeat",
        "name" => "Heartbeat",
        "version" => "1.0.0",
        "author" => "Beep",
        "capabilities" => [ "webhook_ping" ],
        "schedule" => { "default_cron" => "*/15 * * * *" }
      }
    )
    @beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      cron: "*/15 * * * *",
      timezone: "UTC",
      title: "Worker Heartbeat",
      config: { "grace_period_minutes" => 15 }
    )
  end

  test "stamps last_ping_at with valid token" do
    assert_nil @beeper.last_ping_at
    assert @beeper.ping_token.present?

    post api_v1_beeper_app_heartbeat_pings_path(beeper_app_id: "heartbeat", token: @beeper.ping_token)

    assert_response :success
    @beeper.reload
    assert_not_nil @beeper.last_ping_at
  end

  test "returns 404 for unknown token" do
    post api_v1_beeper_app_heartbeat_pings_path(beeper_app_id: "heartbeat", token: "non-existent-token")
    assert_response :not_found
  end

  test "rate limits excessive pings" do
    old_cache = Rails.cache
    Rails.cache = ActiveSupport::Cache::MemoryStore.new
    begin
      60.times do
        post api_v1_beeper_app_heartbeat_pings_path(beeper_app_id: "heartbeat", token: @beeper.ping_token)
        assert_response :success
      end

      post api_v1_beeper_app_heartbeat_pings_path(beeper_app_id: "heartbeat", token: @beeper.ping_token)
      assert_response :too_many_requests
    ensure
      Rails.cache = old_cache
    end
  end
end
