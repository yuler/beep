require "test_helper"

class Api::V1::Channels::Cli::DisconnectsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @account = accounts(:john_account)
    @user = users(:john)
    @channel = @account.channels.create!(
      user: @user,
      kind: :cli,
      name: "laptop"
    )
  end

  test "destroy removes the channel authenticated by CLI token" do
    assert_difference -> { Channel.count }, -1 do
      delete "/api/v1/channels/cli/disconnect",
        headers: { "X-CLI-Token" => @channel.token },
        as: :json
    end

    assert_response :no_content
  end

  test "destroy rejects missing token" do
    delete "/api/v1/channels/cli/disconnect", as: :json

    assert_response :unauthorized
  end

  test "destroy rejects invalid token" do
    delete "/api/v1/channels/cli/disconnect",
      headers: { "X-CLI-Token" => "invalid_token" },
      as: :json

    assert_response :unauthorized
  end
end
