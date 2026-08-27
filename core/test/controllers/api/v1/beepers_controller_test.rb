require "test_helper"

class Api::V1::BeepersControllerTest < ActionDispatch::IntegrationTest
  setup do
    Beeper.seed_official
  end

  test "index returns list of official beepers" do
    get "/api/v1/beepers", as: :json

    assert_response :success
    beepers = response.parsed_body["beepers"]
    assert beepers.size >= 3
    slugs = beepers.map { |b| b["slug"] }
    assert_includes slugs, "site-uptime"
    assert_includes slugs, "ssl-expiry"
    assert_includes slugs, "heartbeat"
  end

  test "show returns details of a single beeper" do
    get "/api/v1/beepers/site-uptime", as: :json

    assert_response :success
    beeper = response.parsed_body
    assert_equal "site-uptime", beeper["slug"]
    assert_equal "Site Uptime & Health Check", beeper["name"]
    assert beeper["inputs"].present?
  end
end
