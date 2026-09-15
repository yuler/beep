require "test_helper"

class Channel::DeliveryTest < ActiveSupport::TestCase
  setup do
    @account = accounts(:john_account)
    @user = users(:john)
    @channel = Channel.create!(
      account: @account,
      user: @user,
      kind: :cli,
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
    succeeded = @channel.deliveries.create!(
      beep_run: @run,
      payload: { title: "Test" }
    )

    assert succeeded.succeed!
    assert_equal "succeeded", succeeded.reload.status
    assert_not succeeded.succeed!, "terminal deliveries must not transition again"

    failed = @channel.deliveries.create!(
      beep_run: @run,
      payload: { title: "Test" }
    )

    assert failed.claim!
    failed.fail!("boom")
    assert_equal "failed", failed.reload.status
    assert_equal "boom", failed.reload.payload["error"]
    assert_not failed.fail!("again"), "terminal deliveries must not transition again"
  end

  test "expire_stale! expires both pending and claimed deliveries past expires_at" do
    pending_stale = @channel.deliveries.create!(
      beep_run: @run,
      payload: { title: "Test pending" },
      expires_at: 1.minute.ago
    )
    claimed_stale = @channel.deliveries.create!(
      beep_run: @run,
      payload: { title: "Test claimed" },
      expires_at: 1.minute.ago
    )
    claimed_stale.update_columns(status: "claimed", claimed_at: 10.minutes.ago)

    Channel::Delivery.expire_stale!

    assert_equal "expired", pending_stale.reload.status
    assert_equal "expired", claimed_stale.reload.status
  end

  test "reclaim_stale! resets unexpired claimed deliveries older than timeout back to pending" do
    stale_claimed = @channel.deliveries.create!(
      beep_run: @run,
      payload: { title: "Test stale claimed" },
      expires_at: 20.minutes.from_now
    )
    stale_claimed.update_columns(status: "claimed", claimed_at: 6.minutes.ago)

    fresh_claimed = @channel.deliveries.create!(
      beep_run: @run,
      payload: { title: "Test fresh claimed" },
      expires_at: 20.minutes.from_now
    )
    fresh_claimed.update_columns(status: "claimed", claimed_at: 1.minute.ago)

    Channel::Delivery.reclaim_stale!

    assert_equal "pending", stale_claimed.reload.status
    assert_nil stale_claimed.claimed_at
    assert_equal "claimed", fresh_claimed.reload.status
  end
end
