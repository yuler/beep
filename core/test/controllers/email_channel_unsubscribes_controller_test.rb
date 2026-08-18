require "test_helper"

class EmailChannelUnsubscribesControllerTest < ActionDispatch::IntegrationTest
  setup do
    @account = accounts(:john_account)
    @token = @account.email_channel_unsubscribe_token
  end

  test "show explains the email switch" do
    get email_channel_unsubscribe_path(@token)

    assert_response :success
    assert_includes response.body, "Turn off email reminders"
    assert @account.reload.email_channel_enabled?
  end

  test "create turns off email reminders without signing in" do
    post email_channel_unsubscribe_path(@token)

    assert_response :success
    assert_not @account.reload.email_channel_enabled?
    assert_includes response.body, "Email reminders are off"
  end

  test "show is not found for a forged token" do
    get email_channel_unsubscribe_path("nope")

    assert_response :not_found
  end
end
