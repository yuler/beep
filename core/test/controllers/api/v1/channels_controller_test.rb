require "test_helper"

class Api::V1::ChannelsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @user = users(:john)
    @session = @identity.sessions.create!
    @token = @session.signed_id
  end

  test "index returns account channels" do
    channel = @account.channels.create!(
      user: @user,
      kind: :cli,
      name: "laptop"
    )

    get "/api/v1/#{@account.slug}/channels",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    channels = response.parsed_body["channels"]
    assert_equal 1, channels.size
    assert_equal channel.id, channels.first["id"]
    assert_equal "laptop", channels.first["name"]
    assert_equal "cli", channels.first["kind"]
    assert_equal @user.id, channels.first.dig("user", "id")
  end

  test "index filters by kind" do
    cli_channel = @account.channels.create!(user: @user, kind: :cli, name: "laptop")
    email_channel = @account.channels.create!(user: @user, kind: :email, name: "john@example.com")

    get "/api/v1/#{@account.slug}/channels?kind=cli",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    channels = response.parsed_body["channels"]
    assert_equal 1, channels.size
    assert_equal cli_channel.id, channels.first["id"]
  end

  test "create creates a channel and returns raw token" do
    assert_difference -> { @account.channels.count }, 1 do
      post "/api/v1/#{@account.slug}/channels",
        params: {
          channel: {
            name: "home-desktop",
            kind: "cli"
          }
        },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :created
    body = response.parsed_body["channel"]
    assert_equal "home-desktop", body["name"]
    assert_equal "cli", body["kind"]
    assert_predicate body["token"], :present?
    assert body["token"].start_with?(Channel::TOKEN_PREFIX)
  end

  test "destroy removes the channel" do
    channel = @account.channels.create!(
      user: @user,
      kind: :cli,
      name: "laptop"
    )

    assert_difference -> { @account.channels.count }, -1 do
      delete "/api/v1/#{@account.slug}/channels/#{channel.id}",
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :no_content
  end

  test "test delivers test notification to channel" do
    channel = @account.channels.create!(
      user: @user,
      kind: :cli,
      name: "laptop"
    )

    assert_difference -> { channel.deliveries.count }, 1 do
      post "/api/v1/#{@account.slug}/channels/#{channel.id}/test",
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :no_content
  end
end
