require "test_helper"

class Api::V1::Cli::AuthorizationsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @session = @identity.sessions.create!
    @token = @session.signed_id
  end

  test "create generates codes and endpoints" do
    post "/api/v1/cli/authorizations",
      params: { client_name: "MacBook Pro" },
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
    auth = Cli::Authorization.create_request!(client_name: "MacBook Pro")

    get "/api/v1/cli/authorizations/#{auth.user_code}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal auth.user_code, body["user_code"]
    assert_equal "MacBook Pro", body["client_name"]
    assert_equal "pending", body["status"]
  end

  test "show returns 404 for invalid code" do
    get "/api/v1/cli/authorizations/INVALID",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
  end

  test "update approves authorization and creates access token" do
    auth = Cli::Authorization.create_request!(client_name: "MacBook Pro")

    assert_difference -> { @identity.access_tokens.count }, 1 do
      patch "/api/v1/cli/authorizations/#{auth.user_code}",
        headers: { "Authorization" => "Bearer #{@token}" },
        params: { client_name: "Workstation" },
        as: :json

      assert_response :success
    end

    assert_equal "approved", auth.reload.status
    assert_equal @identity, auth.identity
    assert_equal "Workstation", auth.access_token.description
  end

  test "destroy denies authorization" do
    auth = Cli::Authorization.create_request!

    delete "/api/v1/cli/authorizations/#{auth.user_code}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :no_content
    assert_equal "access_denied", auth.reload.status
  end

  test "unauthenticated user cannot show or approve" do
    auth = Cli::Authorization.create_request!

    get "/api/v1/cli/authorizations/#{auth.user_code}", as: :json
    assert_response :unauthorized

    patch "/api/v1/cli/authorizations/#{auth.user_code}", as: :json
    assert_response :unauthorized
  end
end
