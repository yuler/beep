require "test_helper"

class Api::V1::PushSubscriptionsControllerTest < ActionDispatch::IntegrationTest
  setup do
    stub_web_push_dns_resolution
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @endpoint = "https://fcm.googleapis.com/fcm/send/abc123"
    @previous_public_key = Rails.application.config.x.vapid.public_key
    @previous_private_key = Rails.application.config.x.vapid.private_key
    Rails.application.config.x.vapid.public_key = "test-vapid-public"
    Rails.application.config.x.vapid.private_key = "test-vapid-private"
  end

  teardown do
    Rails.application.config.x.vapid.public_key = @previous_public_key
    Rails.application.config.x.vapid.private_key = @previous_private_key
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

  test "index lists the current user's subscriptions" do
    mine = users(:john).push_subscriptions.create!(
      endpoint: @endpoint,
      p256dh_key: "key",
      auth_key: "auth"
    )
    users(:yuler).push_subscriptions.create!(
      endpoint: "https://fcm.googleapis.com/fcm/send/other",
      p256dh_key: "key",
      auth_key: "auth"
    )

    get "/api/v1/#{@account.slug}/push_subscriptions",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    ids = response.parsed_body["push_subscriptions"].map { |row| row["id"] }
    assert_equal [ mine.id ], ids
  end

  test "index requires authentication" do
    get "/api/v1/#{@account.slug}/push_subscriptions", as: :json

    assert_response :unauthorized
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

  test "test sends a notification for the current user's subscription" do
    subscription = users(:john).push_subscriptions.create!(
      endpoint: @endpoint,
      p256dh_key: "key",
      auth_key: "auth"
    )
    sent = false

    stub_web_push_payload_send(->(**_kwargs) { sent = true }) do
      post "/api/v1/#{@account.slug}/push_subscriptions/#{subscription.id}/test",
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :no_content
    assert sent
  end

  test "test destroys an expired subscription" do
    subscription = users(:john).push_subscriptions.create!(
      endpoint: @endpoint,
      p256dh_key: "key",
      auth_key: "auth"
    )
    push_response = Object.new
    def push_response.body = "Gone"
    expired = WebPush::ExpiredSubscription.new(push_response, "fcm.googleapis.com")

    stub_web_push_payload_send(->(**_kwargs) { raise expired }) do
      post "/api/v1/#{@account.slug}/push_subscriptions/#{subscription.id}/test",
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :gone
    assert_equal "PUSH_SUBSCRIPTION_EXPIRED", response.parsed_body["code"]
    assert_not Push::Subscription.exists?(subscription.id)
  end

  test "test does not send for another user's subscription" do
    subscription = users(:yuler).push_subscriptions.create!(
      endpoint: @endpoint,
      p256dh_key: "key",
      auth_key: "auth"
    )

    post "/api/v1/#{@account.slug}/push_subscriptions/#{subscription.id}/test",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
  end

  private
    def stub_web_push_payload_send(callable)
      singleton = WebPush.singleton_class
      singleton.alias_method :__orig_payload_send, :payload_send
      singleton.define_method(:payload_send) { |**kwargs| callable.call(**kwargs) }
      yield
    ensure
      singleton.alias_method :payload_send, :__orig_payload_send
      singleton.remove_method :__orig_payload_send
    end
end
