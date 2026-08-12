require "test_helper"

class Api::V1::Admin::JobsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @staff = identities(:yuler)
    @staff.update!(staff: true)
    @token = @staff.sessions.create!.signed_id
  end

  test "show returns adapter info for staff" do
    get api_v1_admin_jobs_url, headers: { "Authorization" => "Bearer #{@token}" }, as: :json

    assert_response :success
    body = response.parsed_body
    assert body["adapter"].present?
    assert_includes [ true, false ], body["available"]
    assert body["recent"].is_a?(Array)
  end

  test "show forbids non-staff" do
    identity = identities(:john)
    identity.update!(staff: false)
    token = identity.sessions.create!.signed_id

    get api_v1_admin_jobs_url, headers: { "Authorization" => "Bearer #{token}" }, as: :json

    assert_response :forbidden
  end
end
