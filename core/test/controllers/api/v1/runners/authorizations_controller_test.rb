require "test_helper"

class Api::V1::Runners::AuthorizationsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @user = users(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
  end

  test "create generates codes and endpoints" do
    post "/api/v1/runners/authorizations",
      params: { runner_name: "My-Server", tags: %w[prod linux], metadata: { os: "linux", arch: "amd64" } },
      as: :json

    assert_response :created
    body = response.parsed_body
    assert_predicate body["device_code"], :present?
    assert_predicate body["user_code"], :present?
    assert_predicate body["verification_uri"], :present?
    assert_predicate body["verification_uri_complete"], :present?
    assert_includes body["verification_uri_complete"], body["user_code"]
    assert_equal 5, body["interval"]
  end

  test "show returns user code details when active" do
    auth = Runner::Authorization.create_request!(
      runner_name: "My-Server",
      tags: %w[prod linux],
      metadata: { "os" => "linux" }
    )

    get "/#{@account.slug}/api/v1/runners/authorizations/#{auth.user_code}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal auth.user_code, body["user_code"]
    assert_equal "My-Server", body["runner_name"]
    assert_equal %w[prod linux], body["tags"]
    assert_equal({ "os" => "linux" }, body["metadata"])
    assert_equal "pending", body["status"]
  end

  test "show returns 404 for invalid code" do
    get "/#{@account.slug}/api/v1/runners/authorizations/INVALID",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
  end

  test "update approves authorization and creates new runner" do
    auth = Runner::Authorization.create_request!(runner_name: "My-Server", tags: %w[default])

    assert_difference -> { Runner.count }, 1 do
      patch "/#{@account.slug}/api/v1/runners/authorizations/#{auth.user_code}",
        params: { runner_name: "Production-Runner", tags: %w[prod gcp] },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :success
    assert_equal "approved", response.parsed_body["status"]
    assert_equal "Production-Runner", response.parsed_body.dig("runner", "name")
    assert_equal %w[prod gcp], response.parsed_body.dig("runner", "tags")
  end

  test "destroy marks authorization as access_denied" do
    auth = Runner::Authorization.create_request!(runner_name: "My-Server")

    delete "/#{@account.slug}/api/v1/runners/authorizations/#{auth.user_code}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :no_content
    assert_equal "access_denied", auth.reload.status
  end

  test "token exchange succeeds after approval" do
    auth = Runner::Authorization.create_request!(runner_name: "My-Server", tags: %w[ci])

    # Poll before approval -> authorization_pending
    post "/api/v1/runners/authorizations/token",
      params: {
        grant_type: "urn:ietf:params:oauth:grant-type:device_code",
        device_code: auth.device_code
      },
      as: :json

    assert_response :bad_request
    assert_equal "authorization_pending", response.parsed_body["error"]

    # Approve
    auth.approve!(user: @user)

    # Poll after interval
    travel 6.seconds do
      post "/api/v1/runners/authorizations/token",
        params: {
          grant_type: "urn:ietf:params:oauth:grant-type:device_code",
          device_code: auth.device_code
        },
        as: :json

      assert_response :success
      body = response.parsed_body
      assert_predicate body["access_token"], :present?
      assert body["access_token"].start_with?("beep_rt_")
      assert_equal "bearer", body["token_type"]
      assert_equal "My-Server", body["runner_name"]
    end
  end
end
