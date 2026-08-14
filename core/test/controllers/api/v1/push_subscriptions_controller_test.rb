require "test_helper"

class Api::V1::PushSubscriptionsControllerTest < ActionDispatch::IntegrationTest
  setup do
    stub_web_push_dns_resolution
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @endpoint = "https://fcm.googleapis.com/fcm/send/abc123"
  end

  test "create stores a subscription for the current user" do
    assert_difference -> { users(:john).push_subscriptions.count }, 1 do
      post "/api/v1/#{@account.slug}/push_subscriptions",
        params: { endpoint: @endpoint, p256dh_key: "key", auth_key: "auth" },
        headers: { "Authorization" => "Bearer #{@token}", "User-Agent" => "Chrome Test" },
        as: :json
    end

    assert_response :created
    body = response.parsed_body
    subscription = users(:john).push_subscriptions.last
    assert_equal subscription.id, body["id"]
    assert_equal @endpoint, body["endpoint"]
    assert_equal "Chrome Test", body["user_agent"]
    assert_equal @account, subscription.account
  end

  test "create is idempotent for the same endpoint" do
    post "/api/v1/#{@account.slug}/push_subscriptions",
      params: { endpoint: @endpoint, p256dh_key: "key", auth_key: "auth" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_no_difference -> { users(:john).push_subscriptions.count } do
      post "/api/v1/#{@account.slug}/push_subscriptions",
        params: { endpoint: @endpoint, p256dh_key: "new-key", auth_key: "new-auth" },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :created
    subscription = users(:john).push_subscriptions.find_by!(endpoint: @endpoint)
    assert_equal "new-key", subscription.p256dh_key
    assert_equal "new-auth", subscription.auth_key
  end

  test "create rejects a non-permitted endpoint" do
    post "/api/v1/#{@account.slug}/push_subscriptions",
      params: { endpoint: "https://attacker.example.com/push", p256dh_key: "key", auth_key: "auth" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert_equal "VALIDATION_ERROR", response.parsed_body["code"]
  end

  test "create requires authentication" do
    post "/api/v1/#{@account.slug}/push_subscriptions",
      params: { endpoint: @endpoint, p256dh_key: "key", auth_key: "auth" },
      as: :json

    assert_response :unauthorized
  end

  test "create returns not found for another account" do
    post "/api/v1/#{accounts(:yuler_account).slug}/push_subscriptions",
      params: { endpoint: @endpoint, p256dh_key: "key", auth_key: "auth" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
  end

  test "destroy deletes the current user's subscription" do
    subscription = users(:john).push_subscriptions.create!(
      endpoint: @endpoint,
      p256dh_key: "key",
      auth_key: "auth"
    )

    assert_difference -> { users(:john).push_subscriptions.count }, -1 do
      delete "/api/v1/#{@account.slug}/push_subscriptions/#{subscription.id}",
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :no_content
  end

  test "destroy does not delete another user's subscription" do
    subscription = users(:yuler).push_subscriptions.create!(
      endpoint: @endpoint,
      p256dh_key: "key",
      auth_key: "auth"
    )

    delete "/api/v1/#{@account.slug}/push_subscriptions/#{subscription.id}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
    assert Push::Subscription.exists?(subscription.id)
  end
end
