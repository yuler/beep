require "test_helper"

class Api::V1::DashboardControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
  end

  test "show returns dashboard stats, chart, and activities" do
    get "/api/v1/#{@account.slug}/dashboard",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body

    assert body["stats"].present?
    assert body["stats"]["executions"].present?
    assert body["stats"]["beeps"].present?
    assert body["stats"]["beepers"].present?
    assert body["stats"]["runners"].present?

    assert body["chart"].present?
    assert_equal "7d", body["chart"]["range"]
    assert_equal 7, body["chart"]["points"].size

    assert body["activities"].is_a?(Array)
  end

  test "show with range=24h returns 24 chart points" do
    get "/api/v1/#{@account.slug}/dashboard",
      params: { range: "24h" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body

    assert_equal "24h", body["chart"]["range"]
    assert_equal 24, body["chart"]["points"].size
  end
end
