require "test_helper"

class Api::V1::SettingsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
  end

  test "show returns the current user's notification channels and timezone" do
    get "/api/v1/#{@account.slug}/settings",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal @account.id, body["id"]
    assert_equal true, body["personal"]
    assert_equal %w[ email web_push ], body["notification_channels"]
    assert_nil body["timezone"]
    assert_nil body["timezone_source"]
  end

  test "update detects timezone only when empty" do
    patch "/api/v1/#{@account.slug}/settings",
      params: { timezone: "Asia/Shanghai", timezone_source: "detected" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "Asia/Shanghai", response.parsed_body["timezone"]
    assert_equal "detected", response.parsed_body["timezone_source"]

    patch "/api/v1/#{@account.slug}/settings",
      params: { timezone: "America/New_York", timezone_source: "detected" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "Asia/Shanghai", response.parsed_body["timezone"]
    assert_equal "detected", users(:john).reload.timezone_source
  end

  test "update manual timezone overwrites detected" do
    users(:john).update!(timezone: "Asia/Shanghai", timezone_source: "detected")

    patch "/api/v1/#{@account.slug}/settings",
      params: { timezone: "Europe/London" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "Europe/London", response.parsed_body["timezone"]
    assert_equal "manual", response.parsed_body["timezone_source"]
  end

  test "update rejects an invalid timezone" do
    patch "/api/v1/#{@account.slug}/settings",
      params: { timezone: "Not/A_Zone" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert_equal "VALIDATION_ERROR", response.parsed_body["code"]
  end


  test "update writes the current user's notification channels" do
    patch "/api/v1/#{@account.slug}/settings",
      params: { notification_channels: %w[ web_push ] },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal %w[ web_push ], response.parsed_body["notification_channels"]
    assert_equal %w[ web_push ], users(:john).reload.notification_channels
  end

  test "update allows an empty notification channel list" do
    patch "/api/v1/#{@account.slug}/settings",
      params: { notification_channels: [] },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal [], response.parsed_body["notification_channels"]
    assert_equal [], users(:john).reload.notification_channels
  end

  test "update writes notification channels on a team account" do
    team = Account.create_with_owner(
      account: { name: "John Team", personal: false, slug: "john_set" },
      owner: { name: "John", identity: @identity }
    )
    owner = team.users.find_by!(role: :owner)

    patch "/api/v1/#{team.slug}/settings",
      params: { notification_channels: %w[ email ] },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal %w[ email ], response.parsed_body["notification_channels"]
    assert_equal %w[ email ], owner.reload.notification_channels
    assert_equal %w[ email web_push ], users(:john).reload.notification_channels
  end

  test "show requires authentication" do
    get "/api/v1/#{@account.slug}/settings", as: :json

    assert_response :unauthorized
  end
end
