require "test_helper"

class Api::V1::DeviceControllerTest < ActionDispatch::IntegrationTest
  setup do
    @account = accounts(:john_account)
    @user = users(:john)
    @channel = @account.channels.create!(
      user: @user,
      kind: :device,
      name: "laptop"
    )
    @beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Get off work",
      run_at: 1.hour.from_now
    )
    @run = @beep.runs.create!(scheduled_for: Time.current, status: :pending)
    @delivery = @channel.deliveries.create!(
      beep_run: @run,
      payload: { title: "Get off work" }
    )
  end

  test "inbox returns pending deliveries and claims them" do
    get "/api/v1/device/inbox",
      headers: { "X-Device-Token" => @channel.token },
      as: :json

    assert_response :success
    deliveries = response.parsed_body["deliveries"]
    assert_equal 1, deliveries.size
    assert_equal @delivery.id, deliveries.first["id"]
    assert_equal "Get off work", deliveries.first.dig("payload", "title")

    # Delivery should now be claimed
    assert_equal "claimed", @delivery.reload.status
  end

  test "inbox rejects unauthorized request" do
    get "/api/v1/device/inbox",
      headers: { "X-Device-Token" => "invalid_token" },
      as: :json

    assert_response :unauthorized
  end

  test "ack marks delivery succeeded" do
    post "/api/v1/device/deliveries/#{@delivery.id}/ack",
      params: { status: "succeeded" },
      headers: { "X-Device-Token" => @channel.token },
      as: :json

    assert_response :no_content
    assert_equal "succeeded", @delivery.reload.status
  end

  test "ack marks delivery failed" do
    post "/api/v1/device/deliveries/#{@delivery.id}/ack",
      params: { status: "failed", error: "Execution error" },
      headers: { "X-Device-Token" => @channel.token },
      as: :json

    assert_response :no_content
    assert_equal "failed", @delivery.reload.status
  end
end
