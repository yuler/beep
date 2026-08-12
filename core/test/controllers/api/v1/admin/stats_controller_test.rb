require "test_helper"

class Api::V1::Admin::StatsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @staff = identities(:yuler)
    @staff.update!(staff: true)
    @session = @staff.sessions.create!
    @token = @session.signed_id
  end

  test "show returns account and identity stats for staff" do
    get api_v1_admin_stats_url, headers: { "Authorization" => "Bearer #{@token}" }, as: :json

    assert_response :success
    body = response.parsed_body
    assert body["accounts"]["total"] >= 1
    assert body["identities"]["total"] >= 1
    assert body["recent_accounts"].is_a?(Array)
  end

  test "show forbids non-staff" do
    identity = identities(:john)
    identity.update!(staff: false)
    token = identity.sessions.create!.signed_id

    get api_v1_admin_stats_url, headers: { "Authorization" => "Bearer #{token}" }, as: :json

    assert_response :forbidden
  end
end
