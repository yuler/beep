require "test_helper"

class ChannelDeliveryTest < ActiveSupport::TestCase
  setup do
    @account = accounts(:john_account)
    @user = users(:john)
    @channel = Channel.create!(
      account: @account,
      user: @user,
      kind: :device,
      name: "laptop"
    )
    @beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Test beep",
      run_at: 1.hour.from_now
    )
    @run = @beep.runs.create!(scheduled_for: Time.current, status: :pending)
  end

  test "creates delivery with default expires_at" do
    delivery = @channel.deliveries.create!(
      beep_run: @run,
      payload: { title: "Test" }
    )

    assert_equal "pending", delivery.status
    assert_predicate delivery.expires_at, :present?
    assert delivery.expires_at > Time.current
    assert_not delivery.expired?
  end

  test "claim! transitions pending to claimed" do
    delivery = @channel.deliveries.create!(
      beep_run: @run,
      payload: { title: "Test" }
    )

    assert delivery.claim!
    assert_equal "claimed", delivery.reload.status
    assert_predicate delivery.claimed_at, :present?
  end

  test "claim! fails if delivery is expired" do
    delivery = @channel.deliveries.create!(
      beep_run: @run,
      payload: { title: "Test" },
      expires_at: 10.minutes.ago
    )

    assert delivery.expired?
    assert_not delivery.claim!
    assert_equal "pending", delivery.reload.status
  end

  test "succeed! and fail! transitions" do
    delivery = @channel.deliveries.create!(
      beep_run: @run,
      payload: { title: "Test" }
    )

    delivery.succeed!
    assert_equal "succeeded", delivery.reload.status

    delivery.fail!
    assert_equal "failed", delivery.reload.status
  end
end
