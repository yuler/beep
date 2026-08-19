require "test_helper"

class EmailChannelUnsubscribesControllerTest < ActionDispatch::IntegrationTest
  setup do
    @user = users(:john)
    @token = @user.email_channel_unsubscribe_token
  end

  test "show explains the email switch" do
    get email_channel_unsubscribe_path(@token)

    assert_response :success
    assert_includes response.body, "Turn off email reminders"
    assert_includes @user.reload.notification_channels, "email"
  end

  test "create removes email from that user without signing in" do
    post email_channel_unsubscribe_path(@token)

    assert_response :success
    assert_not_includes @user.reload.notification_channels, "email"
    assert_includes @user.notification_channels, "web_push"
    assert_includes response.body, "Email reminders are off"
  end

  test "show is not found for a forged token" do
    get email_channel_unsubscribe_path("nope")

    assert_response :not_found
  end
end
