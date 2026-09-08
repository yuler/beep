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

  test "reclaim does not fail a run that was claimed after pending check" do
    job = @runner.jobs.create!(name: "Check", slug: "check", cron: "*/5 * * * *", timezone: "UTC")
    job.update_columns(status: "firing", updated_at: 5.minutes.ago)
    run = job.runs.create!(scheduled_for: 3.minutes.ago, status: "pending", runner: @runner, created_at: 3.minutes.ago)

    assert run.claim_for(@runner)
    job.reclaim_stale

    run.reload
    assert_equal "running", run.status
    assert_nil run.result_status
  end

  test "reclaim preserves pending runs when runner is online" do
    job = @runner.jobs.create!(name: "Check", slug: "check", cron: "*/5 * * * *", timezone: "UTC")
    job.update_columns(status: "firing", updated_at: 5.minutes.ago)
    run = job.runs.create!(scheduled_for: 5.minutes.ago, status: "pending", runner: @runner, created_at: 5.minutes.ago)

    job.reclaim_stale

    run.reload
    job.reload
    assert_equal "pending", run.status
    assert job.firing?
  end

  test "reclaim fails pending runs when runner is offline" do
    @runner.update_columns(status: "offline", last_seen_at: 3.minutes.ago)
    job = @runner.jobs.create!(name: "Check", slug: "check", cron: "*/5 * * * *", timezone: "UTC")
    job.update_columns(status: "firing", updated_at: 5.minutes.ago)
    run = job.runs.create!(scheduled_for: 5.minutes.ago, status: "pending", runner: @runner, created_at: 5.minutes.ago)

    job.reclaim_stale

    run.reload
    job.reload
    assert_equal "failed", run.status
    assert_equal "error", run.result_status
    assert_equal "Runner offline", run.result["title"]
    assert job.active?
  end

  test "reclaim_stale_firing recovers open runs on paused jobs" do
    job = @runner.jobs.create!(name: "Check", slug: "check", cron: "*/5 * * * *", timezone: "UTC")
    job.update_columns(status: "firing")
    run = job.runs.create!(scheduled_for: 5.minutes.ago, status: "pending", runner: @runner, created_at: 5.minutes.ago)
    job.pause!

    Runner::Job.reclaim_stale_firing

    run.reload
    job.reload
    assert_equal "failed", run.status
    assert_equal "error", run.result_status
    assert job.paused?
  end

  test "timezone change refreshes next_run_at" do
    job = @runner.jobs.create!(name: "Check", slug: "check", cron: "0 9 * * *", timezone: "UTC")
    original = job.next_run_at

    job.update!(timezone: "Asia/Shanghai")

    assert_not_equal original, job.next_run_at
  end

  test "record_result! is atomic and rejects second write" do
    job = @runner.jobs.create!(name: "Check", slug: "check", cron: "*/5 * * * *", timezone: "UTC")
    job.update_columns(status: "firing")
    run = job.runs.create!(scheduled_for: Time.current, status: "running", runner: @runner, claimed_at: Time.current)

    assert run.record_result!(status: :ok, title: "first", run_status: :succeeded)
    refute run.record_result!(status: :error, title: "second", run_status: :failed)

    run.reload
    assert_equal "succeeded", run.status
    assert_equal "ok", run.result_status
    assert_equal "first", run.result["title"]
  end
end
