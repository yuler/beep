require "test_helper"

class Runner::JobTest < ActiveSupport::TestCase
  setup do
    @account = accounts(:john_account)
    @runner = @account.runners.create!(name: "Office")
    @runner.update_columns(status: "online", last_seen_at: 5.seconds.ago)
  end

  test "normalizes slug and assigns account from runner" do
    job = @runner.jobs.create!(
      name: "Intranet HTTP",
      slug: " Intranet-HTTP ",
      cron: "*/5 * * * *",
      timezone: "UTC"
    )

    assert_equal "intranet-http", job.slug
    assert_equal @account.id, job.account_id
    assert_not_nil job.next_run_at
  end

  test "rejects invalid slug" do
    job = @runner.jobs.new(name: "Bad", slug: "Nope Space", cron: "*/5 * * * *", timezone: "UTC")
    assert_not job.valid?
    assert_includes job.errors[:slug], "is invalid"
  end

  test "poll_due_now creates a pending run when runner is online" do
    job = @runner.jobs.create!(name: "Check", slug: "check", cron: "*/5 * * * *", timezone: "UTC")
    job.update_columns(next_run_at: 1.minute.ago, status: "active")

    Runner::Job.poll_due_now

    job.reload
    assert job.firing?
    assert_equal 1, job.runs.count
    assert_equal "pending", job.runs.first.status
  end

  test "poll_due_now records error when runner is offline" do
    @runner.update_columns(status: "offline", last_seen_at: 2.minutes.ago)
    job = @runner.jobs.create!(name: "Check", slug: "check", cron: "*/5 * * * *", timezone: "UTC")
    job.update_columns(next_run_at: 1.minute.ago, status: "active")

    Runner::Job.poll_due_now

    job.reload
    assert job.active?
    run = job.runs.first
    assert_equal "failed", run.status
    assert_equal "error", run.result_status
    assert_equal "Runner offline", run.result["title"]
  end
end
