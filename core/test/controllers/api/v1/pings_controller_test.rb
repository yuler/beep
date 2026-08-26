require "test_helper"

class Api::V1::PingsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @account = accounts(:john_account)
    @plugin = Plugin.create!(
      slug: "heartbeat",
      version: "1.0.0",
      manifest: {
        "manifest_version" => 1,
        "slug" => "heartbeat",
        "name" => "Heartbeat",
        "version" => "1.0.0",
        "author" => "Beep",
        "ingest" => { "webhook" => true },
        "schedule" => { "default_cron" => "*/15 * * * *" }
      }
    )
    @beep = Beep.create!(
      account: @account,
      plugin: @plugin,
      kind: :recurring,
      cron: "*/15 * * * *",
      timezone: "UTC",
      title: "Worker Heartbeat",
      plugin_config: { "grace_period_minutes" => 15 }
    )
  end

  test "stamps last_ping_at with valid token" do
    assert_nil @beep.last_ping_at
    assert @beep.ping_token.present?

    post "/api/v1/ping/#{@beep.ping_token}"

    assert_response :success
    @beep.reload
    assert_not_nil @beep.last_ping_at
  end

  test "returns 404 for unknown token" do
    post "/api/v1/ping/non-existent-token"
    assert_response :not_found
  end
end
