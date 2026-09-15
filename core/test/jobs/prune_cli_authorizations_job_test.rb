require "test_helper"

class PruneCliAuthorizationsJobTest < ActiveJob::TestCase
  test "expires pending authorizations past expires_at" do
    expired = Cli::Authorization.create_request!(client_name: "Old")
    expired.update_columns(expires_at: 1.minute.ago)
    fresh = Cli::Authorization.create_request!(client_name: "New")

    PruneCliAuthorizationsJob.perform_now

    assert_equal "expired", expired.reload.status
    assert_equal "pending", fresh.reload.status
  end

  test "leaves non-pending authorizations untouched" do
    auth = Cli::Authorization.create_request!(client_name: "Denied")
    auth.update_columns(status: "access_denied", expires_at: 1.minute.ago)

    PruneCliAuthorizationsJob.perform_now

    assert_equal "access_denied", auth.reload.status
  end
end
