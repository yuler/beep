require "test_helper"

class Api::V1::PluginsControllerTest < ActionDispatch::IntegrationTest
  setup do
    Plugin.seed_official_plugins!
  end

  test "index returns list of official plugins" do
    get "/api/v1/plugins", as: :json

    assert_response :success
    plugins = response.parsed_body["plugins"]
    assert plugins.size >= 3
    slugs = plugins.map { |p| p["slug"] }
    assert_includes slugs, "site-uptime"
    assert_includes slugs, "ssl-expiry"
    assert_includes slugs, "heartbeat"
  end

  test "show returns details of a single plugin" do
    get "/api/v1/plugins/site-uptime", as: :json

    assert_response :success
    plugin = response.parsed_body
    assert_equal "site-uptime", plugin["slug"]
    assert_equal "Site Uptime & Health Check", plugin["name"]
    assert plugin["inputs"].present?
  end
end
