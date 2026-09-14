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
end
