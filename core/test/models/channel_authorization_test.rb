require "test_helper"

class ChannelAuthorizationTest < ActiveSupport::TestCase
  setup do
    @user = users(:john)
  end

  test "generates device_code and user_code on create" do
    auth = ChannelAuthorization.create_request!(channel_name: "MacBook")
    assert_not_nil auth.device_code
    assert_not_nil auth.user_code
    assert_match /\A[BCDFGHJKMNPQRSTVWXYZ23456789]{4}-[BCDFGHJKMNPQRSTVWXYZ23456789]{4}\z/, auth.user_code
    assert_equal "pending", auth.status
    assert auth.expires_at > Time.current
  end

  test "approve! creates a new CLI channel and updates status" do
    auth = ChannelAuthorization.create_request!(channel_name: "Work-Laptop")
    assert_difference -> { Channel.count }, 1 do
      success = auth.approve!(user: @user, name: "Custom Name")
      assert success
    end

    assert_equal "approved", auth.reload.status
    assert_equal "Custom Name", auth.channel.name
    assert_equal "cli", auth.channel.kind
    assert_equal @user, auth.channel.user
    assert_equal @user.account, auth.channel.account
    assert_equal @user.account, auth.account
    assert_match /\Abeep_ct_/, auth.channel.token
  end

  test "deny! updates status to access_denied" do
    auth = ChannelAuthorization.create_request!(channel_name: "Work-Laptop")
    auth.deny!
    assert_equal "access_denied", auth.reload.status
  end

  test "poll! expires pending authorization after expires_at" do
    auth = ChannelAuthorization.create_request!(channel_name: "Work-Laptop")
    auth.update_columns(expires_at: 1.minute.ago)
    auth.poll!
    assert_equal "expired", auth.reload.status
  end
end
