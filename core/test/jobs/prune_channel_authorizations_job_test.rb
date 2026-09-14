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
end
