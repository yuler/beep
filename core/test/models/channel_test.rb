require "test_helper"

class ChannelTest < ActiveSupport::TestCase
  setup do
    @account = accounts(:john_account)
    @user = users(:john)
  end

  test "generates secure token on creation" do
    channel = Channel.create!(
      account: @account,
      user: @user,
      kind: :cli,
      name: "my-laptop"
    )

    assert_predicate channel.token, :present?
    assert channel.token.start_with?(Channel::TOKEN_PREFIX)
    assert_equal "active", channel.status
  end

  test "validates user belongs to account" do
    other_account = accounts(:yuler_account)
    channel = Channel.new(
      account: other_account,
      user: @user,
      kind: :cli,
      name: "laptop"
    )

    assert_not channel.valid?
    assert_includes channel.errors[:user], "must belong to the same account"
  end

  test "touch_last_seen updates last_seen_at" do
    channel = Channel.create!(
      account: @account,
      user: @user,
      kind: :cli,
      name: "my-laptop"
    )

    assert_nil channel.last_seen_at
    channel.touch_last_seen
    assert_not_nil channel.reload.last_seen_at
  end

  test "masked_token masks sensitive token" do
    channel = Channel.create!(
      account: @account,
      user: @user,
      kind: :cli,
      name: "my-laptop"
    )

    assert_equal "#{Channel::TOKEN_PREFIX}••••", channel.masked_token
  end

  test "destroying channel cascades to dependent authorizations" do
    auth = Channel::Authorization.create_request!
    auth.approve!(user: @user)
    channel = auth.channel

    assert_difference -> { Channel::Authorization.count }, -1 do
      channel.destroy!
    end
  end
end
