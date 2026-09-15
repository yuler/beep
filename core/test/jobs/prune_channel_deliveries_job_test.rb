require "test_helper"

class PruneChannelDeliveriesJobTest < ActiveJob::TestCase
  setup do
    @channel = Channel.create!(
      account: accounts(:john_account),
      user: users(:john),
      kind: :cli,
      name: "test-laptop"
    )
  end

  test "expires pending and claimed deliveries past expires_at" do
    pending_stale = @channel.deliveries.create!(expires_at: 1.minute.ago)
    claimed_stale = @channel.deliveries.create!(expires_at: 1.minute.ago)
    claimed_stale.update_columns(status: "claimed", claimed_at: 10.minutes.ago)
    fresh = @channel.deliveries.create!(expires_at: 10.minutes.from_now)

    PruneChannelDeliveriesJob.perform_now

    assert_equal "expired", pending_stale.reload.status
    assert_equal "expired", claimed_stale.reload.status
    assert_equal "pending", fresh.reload.status
  end

  test "prunes terminal deliveries older than retention period" do
    old_delivery = @channel.deliveries.create!(status: "succeeded")
    old_delivery.update_columns(updated_at: 31.days.ago)

    recent_delivery = @channel.deliveries.create!(status: "succeeded")
    recent_delivery.update_columns(updated_at: 1.day.ago)

    PruneChannelDeliveriesJob.perform_now

    assert_not Channel::Delivery.exists?(old_delivery.id)
    assert Channel::Delivery.exists?(recent_delivery.id)
  end
end
