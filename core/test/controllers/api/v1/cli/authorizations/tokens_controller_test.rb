require "test_helper"

class Api::V1::Cli::Authorizations::TokensControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
  end

  test "returns authorization_pending when pending" do
    auth = Cli::Authorization.create_request!

    post "/api/v1/cli/authorizations/token",
      params: {
        grant_type: "urn:ietf:params:oauth:grant-type:device_code",
        device_code: auth.device_code
      },
      as: :json

    assert_response :bad_request
    assert_equal "authorization_pending", response.parsed_body["error"]
  end

  test "returns token when approved" do
    auth = Cli::Authorization.create_request!(client_name: "MacBook Pro")
    auth.approve!(identity: @identity)

    post "/api/v1/cli/authorizations/token",
      params: {
        grant_type: "urn:ietf:params:oauth:grant-type:device_code",
        device_code: auth.device_code
      },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal "bearer", body["token_type"]
    assert_predicate body["access_token"], :present?
    assert body["access_token"].start_with?("beep_pat_")
    assert_equal @identity.id, body["user"]["id"]
    assert_equal @identity.email, body["user"]["email"]
    assert_equal "consumed", auth.reload.status
  end

  test "returns access_denied when denied" do
    auth = Cli::Authorization.create_request!
    auth.deny!

    post "/api/v1/cli/authorizations/token",
      params: {
        grant_type: "urn:ietf:params:oauth:grant-type:device_code",
        device_code: auth.device_code
      },
      as: :json

    assert_response :bad_request
    assert_equal "access_denied", response.parsed_body["error"]
  end

  test "returns expired_token when approved but expired" do
    auth = Cli::Authorization.create_request!
    auth.approve!(identity: @identity)
    auth.update_columns(expires_at: 1.minute.ago)

    post "/api/v1/cli/authorizations/token",
      params: {
        grant_type: "urn:ietf:params:oauth:grant-type:device_code",
        device_code: auth.device_code
      },
      as: :json

    assert_response :bad_request
    assert_equal "expired_token", response.parsed_body["error"]
    assert_equal "expired", auth.reload.status
  end

  test "returns expired_token when expired" do
    auth = Cli::Authorization.create_request!
    auth.update_columns(expires_at: 1.minute.ago)

    post "/api/v1/cli/authorizations/token",
      params: {
        grant_type: "urn:ietf:params:oauth:grant-type:device_code",
        device_code: auth.device_code
      },
      as: :json

    assert_response :bad_request
    assert_equal "expired_token", response.parsed_body["error"]
  end

  test "rejects invalid grant type" do
    auth = Cli::Authorization.create_request!

    post "/api/v1/cli/authorizations/token",
      params: {
        grant_type: "password",
        device_code: auth.device_code
      },
      as: :json

    assert_response :bad_request
    assert_equal "unsupported_grant_type", response.parsed_body["error"]
  end

  test "rejects invalid device code" do
    post "/api/v1/cli/authorizations/token",
      params: {
        grant_type: "urn:ietf:params:oauth:grant-type:device_code",
        device_code: "nonexistent"
      },
      as: :json

    assert_response :bad_request
    assert_equal "invalid_grant", response.parsed_body["error"]
  end
end
