require "test_helper"

class PruneRunnerAuthorizationsJobTest < ActiveJob::TestCase
  test "expires pending authorizations past expires_at" do
    expired = Runner::Authorization.create_request!(runner_name: "Old")
    expired.update_columns(expires_at: 1.minute.ago)
    fresh = Runner::Authorization.create_request!(runner_name: "New")

    PruneRunnerAuthorizationsJob.perform_now

    assert_equal "expired", expired.reload.status
    assert_equal "pending", fresh.reload.status
  end

  test "leaves non-pending authorizations untouched" do
    auth = Runner::Authorization.create_request!(runner_name: "Denied")
    auth.update_columns(status: "access_denied", expires_at: 1.minute.ago)

    PruneRunnerAuthorizationsJob.perform_now

    assert_equal "access_denied", auth.reload.status
  end

  test "prunes terminal authorizations older than retention period" do
    old_auth = Runner::Authorization.create_request!(runner_name: "VeryOld")
    old_auth.update_columns(status: "expired", updated_at: 8.days.ago)

    recent_expired = Runner::Authorization.create_request!(runner_name: "Recent")
    recent_expired.update_columns(status: "expired", updated_at: 1.day.ago)

    PruneRunnerAuthorizationsJob.perform_now

    assert_not Runner::Authorization.exists?(old_auth.id)
    assert Runner::Authorization.exists?(recent_expired.id)
  end
end
