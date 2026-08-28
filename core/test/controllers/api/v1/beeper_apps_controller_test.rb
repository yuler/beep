require "test_helper"

class Api::V1::BeeperAppsControllerTest < ActionDispatch::IntegrationTest
  setup do
    BeeperApp.seed_official
  end

  test "index returns list of official beeper apps" do
    get "/api/v1/beeper_apps", as: :json

    assert_response :success
    beeper_apps = response.parsed_body["beeper_apps"]
    assert beeper_apps.size >= 3
    slugs = beeper_apps.map { |b| b["slug"] }
    assert_includes slugs, "site-uptime"
    assert_includes slugs, "ssl-expiry"
    assert_includes slugs, "heartbeat"
  end

  test "show returns details of a single beeper app" do
    get "/api/v1/beeper_apps/site-uptime", as: :json

    assert_response :success
    beeper_app = response.parsed_body
    assert_equal "site-uptime", beeper_app["slug"]
    assert_equal "Site Uptime & Health Check", beeper_app["name"]
    assert beeper_app["inputs"].present?
  end
end
