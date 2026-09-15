require "test_helper"

class PruneChannelAuthorizationsJobTest < ActiveJob::TestCase
  test "expires pending authorizations past expires_at" do
    expired = Channel::Authorization.create_request!(channel_name: "Old")
    expired.update_columns(expires_at: 1.minute.ago)
    fresh = Channel::Authorization.create_request!(channel_name: "New")

    PruneChannelAuthorizationsJob.perform_now

    assert_equal "expired", expired.reload.status
    assert_equal "pending", fresh.reload.status
  end

  test "leaves non-pending authorizations untouched" do
    auth = Channel::Authorization.create_request!(channel_name: "Denied")
    auth.update_columns(status: "access_denied", expires_at: 1.minute.ago)

    PruneChannelAuthorizationsJob.perform_now

    assert_equal "access_denied", auth.reload.status
  end

  test "prunes terminal authorizations and deliveries older than retention period" do
    old_auth = Channel::Authorization.create_request!(channel_name: "VeryOld")
    old_auth.update_columns(status: "expired", updated_at: 8.days.ago)

    recent_expired = Channel::Authorization.create_request!(channel_name: "Recent")
    recent_expired.update_columns(status: "expired", updated_at: 1.day.ago)

    channel = Channel.create!(account: accounts(:john_account), user: users(:john), kind: :cli, name: "test-laptop")
    old_delivery = channel.deliveries.create!(status: "succeeded")
    old_delivery.update_columns(updated_at: 31.days.ago)

    recent_delivery = channel.deliveries.create!(status: "succeeded")
    recent_delivery.update_columns(updated_at: 1.day.ago)

    PruneChannelAuthorizationsJob.perform_now

    assert_not Channel::Authorization.exists?(old_auth.id)
    assert Channel::Authorization.exists?(recent_expired.id)
    assert_not Channel::Delivery.exists?(old_delivery.id)
    assert Channel::Delivery.exists?(recent_delivery.id)
  end
end
