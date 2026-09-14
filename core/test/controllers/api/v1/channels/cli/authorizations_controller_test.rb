require "test_helper"

class Api::V1::Channels::Cli::AuthorizationsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @user = users(:john)
    @session = @identity.sessions.create!
    @token = @session.signed_id
  end

  test "create generates codes and endpoints" do
    post "/api/v1/channels/cli/authorizations",
      params: { channel_name: "My-Laptop" },
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

  test "create with account_slug generates account-scoped endpoints" do
    post "/api/v1/channels/cli/authorizations",
      params: { channel_name: "My-Laptop", account_slug: "acme" },
      as: :json

    assert_response :created
    body = response.parsed_body
    assert_includes body["verification_uri"], "/acme/device/channel"
    assert_includes body["verification_uri_complete"], "/acme/device/channel?code="
  end

  test "show returns user code details when active" do
    auth = Channel::Authorization.create_request!(channel_name: "My-Laptop")

    get "/api/v1/channels/cli/authorizations/#{auth.user_code}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal auth.user_code, body["user_code"]
    assert_equal "My-Laptop", body["channel_name"]
    assert_equal "pending", body["status"]
  end

  test "show returns 404 for invalid code" do
    get "/api/v1/channels/cli/authorizations/INVALID",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
  end

  test "show returns not found when user is not member of target account" do
    auth = Channel::Authorization.create_request!(channel_name: "My-Laptop")
    other_account = accounts(:yuler_account)

    get "/api/v1/#{other_account.slug}/channels/cli/authorizations/#{auth.user_code}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
  end

  test "update approves authorization and creates new channel" do
    auth = Channel::Authorization.create_request!(channel_name: "My-Laptop")

    assert_difference -> { Channel.count }, 1 do
      patch "/api/v1/channels/cli/authorizations/#{auth.user_code}",
        params: { channel_name: "Renamed-Laptop" },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :success
    assert_equal "approved", response.parsed_body["status"]
    assert_equal "Renamed-Laptop", response.parsed_body.dig("channel", "name")
    assert_equal "cli", response.parsed_body.dig("channel", "kind")

    assert_equal "approved", auth.reload.status
    assert_equal @user.account, auth.account
    assert_equal "Renamed-Laptop", auth.channel.name
  end

  test "update approves authorization under account slug path" do
    auth = Channel::Authorization.create_request!(channel_name: "My-Laptop")

    assert_difference -> { Channel.count }, 1 do
      patch "/api/v1/#{@user.account.slug}/channels/cli/authorizations/#{auth.user_code}",
        params: { channel_name: "Team-Laptop" },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :success
    assert_equal "approved", response.parsed_body["status"]
    assert_equal "Team-Laptop", response.parsed_body.dig("channel", "name")

    assert_equal "approved", auth.reload.status
    assert_equal @user.account, auth.account
    assert_equal "Team-Laptop", auth.channel.name
  end

  test "update returns not found when user is not member of target account" do
    auth = Channel::Authorization.create_request!(channel_name: "My-Laptop")
    other_account = accounts(:yuler_account)

    patch "/api/v1/#{other_account.slug}/channels/cli/authorizations/#{auth.user_code}",
      params: { channel_name: "Hacked-Laptop" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
    assert_equal "pending", auth.reload.status
  end

  test "update returns unprocessable entity when channel name exceeds maximum length" do
    auth = Channel::Authorization.create_request!(channel_name: "My-Laptop")

    patch "/api/v1/#{@user.account.slug}/channels/cli/authorizations/#{auth.user_code}",
      params: { channel_name: "A" * 81 },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert_equal "VALIDATION_ERROR", response.parsed_body["code"]
    assert_includes response.parsed_body["message"], "too long"
    assert_equal "pending", auth.reload.status
  end

  test "destroy denies authorization under account slug path" do
    auth = Channel::Authorization.create_request!(channel_name: "My-Laptop")

    delete "/api/v1/#{@user.account.slug}/channels/cli/authorizations/#{auth.user_code}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :no_content
    assert_equal "access_denied", auth.reload.status
  end

  test "destroy returns not found when user is not member of target account" do
    auth = Channel::Authorization.create_request!(channel_name: "My-Laptop")
    other_account = accounts(:yuler_account)

    delete "/api/v1/#{other_account.slug}/channels/cli/authorizations/#{auth.user_code}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
    assert_equal "pending", auth.reload.status
  end

  test "tokens#create polling returns authorization_pending when pending" do
    auth = Channel::Authorization.create_request!(channel_name: "My-Laptop")

    post "/api/v1/channels/cli/authorizations/token",
      params: {
        grant_type: "urn:ietf:params:oauth:grant-type:device_code",
        device_code: auth.device_code
      },
      as: :json

    assert_response :bad_request
    assert_equal "authorization_pending", response.parsed_body["error"]
  end

  test "tokens#create polling returns access_token when approved" do
    auth = Channel::Authorization.create_request!(channel_name: "My-Laptop")
    auth.approve!(user: @user, name: "My-Laptop")

    post "/api/v1/channels/cli/authorizations/token",
      params: {
        grant_type: "urn:ietf:params:oauth:grant-type:device_code",
        device_code: auth.device_code
      },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_predicate body["access_token"], :present?
    assert body["access_token"].start_with?(Channel::TOKEN_PREFIX)
    assert_equal "bearer", body["token_type"]
    assert_equal auth.channel.id, body["channel_id"]
    assert_equal "My-Laptop", body["channel_name"]
  end

  test "tokens#create polling returns access_denied when denied" do
    auth = Channel::Authorization.create_request!(channel_name: "My-Laptop")
    auth.deny!

    post "/api/v1/channels/cli/authorizations/token",
      params: {
        grant_type: "urn:ietf:params:oauth:grant-type:device_code",
        device_code: auth.device_code
      },
      as: :json

    assert_response :bad_request
    assert_equal "access_denied", response.parsed_body["error"]
  end

  test "tokens#create polling returns expired_token when expired" do
    auth = Channel::Authorization.create_request!(channel_name: "My-Laptop")
    auth.update_columns(expires_at: 1.minute.ago)

    post "/api/v1/channels/cli/authorizations/token",
      params: {
        grant_type: "urn:ietf:params:oauth:grant-type:device_code",
        device_code: auth.device_code
      },
      as: :json

    assert_response :bad_request
    assert_equal "expired_token", response.parsed_body["error"]
  end
end
