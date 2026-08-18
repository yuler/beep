require "test_helper"

class Api::V1::SettingsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
  end

  test "show returns personal email channel switch" do
    get "/api/v1/#{@account.slug}/settings",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal @account.id, body["id"]
    assert_equal true, body["personal"]
    assert_equal true, body["email_channel_enabled"]
  end

  test "update turns off email reminders" do
    patch "/api/v1/#{@account.slug}/settings",
      params: { email_channel_enabled: false },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal false, response.parsed_body["email_channel_enabled"]
    assert_not @account.reload.email_channel_enabled?
  end

  test "update rejects email switch on a team account" do
    team = Account.create_with_owner(
      account: { name: "John Team", personal: false, slug: "john_set" },
      owner: { name: "John", identity: @identity }
    )

    patch "/api/v1/#{team.slug}/settings",
      params: { email_channel_enabled: false },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert team.reload.email_channel_enabled?
  end

  test "show requires authentication" do
    get "/api/v1/#{@account.slug}/settings", as: :json

    assert_response :unauthorized
  end
end
