require "test_helper"

class BeeperPollerTest < ActiveSupport::TestCase
  include ActiveJob::TestHelper

  setup do
    @account = accounts(:john_account)
    @beeper_app = BeeperApp.create!(slug: "echo", version: "1.0.0", manifest: echo_manifest)
  end

  test "poll_due_now claims an active beeper whose next_run_at is past" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Echo Beeper",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "status" => "ok" },
      notification_channels: %w[ email ]
    )
    beeper.update_columns(next_run_at: 1.minute.ago, status: "active")

    Beeper.poll_due_now

    beeper.reload
    assert beeper.firing? || beeper.active?
    assert_equal 1, beeper.runs.count
    assert_equal "pending", beeper.runs.first.status
  end

  test "poll_due_now ignores paused and non-due beepers" do
    paused_beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Paused Echo",
      cron: "*/5 * * * *",
      timezone: "UTC",
      status: "paused",
      next_run_at: 1.minute.ago,
      notification_channels: %w[ email ]
    )
    future_beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Future Echo",
      cron: "*/5 * * * *",
      timezone: "UTC",
      next_run_at: 5.minutes.from_now,
      notification_channels: %w[ email ]
    )

    Beeper.poll_due_now

    assert_equal 0, paused_beeper.runs.count
    assert_equal 0, future_beeper.runs.count
  end

  test "pause! and resume! change beeper status and calculate next_run_at" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Echo Lifecycle",
      cron: "*/5 * * * *",
      timezone: "UTC",
      notification_channels: %w[ email ]
    )
    assert beeper.active?
    assert_not_nil beeper.next_run_at

    beeper.pause!
    assert beeper.paused?

    beeper.resume!
    assert beeper.active?
    assert_not_nil beeper.next_run_at
  end

  test "default notification_channels copies from account owner if empty" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Default Channels",
      cron: "*/5 * * * *",
      timezone: "UTC"
    )
    assert_equal @account.owner_user.notification_channels, beeper.notification_channels
  end

  test "claim_run re-enqueues job when duplicate pending run exists" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Duplicate Run",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "status" => "ok" },
      notification_channels: %w[ email ]
    )
    scheduled_for = 1.minute.from_now.change(usec: 0)
    beeper.update_columns(next_run_at: scheduled_for, status: "active")
    existing = beeper.runs.create!(scheduled_for: scheduled_for, status: :pending)

    assert_enqueued_with(job: RunBeeperJob, args: [ existing ]) do
      beeper.send(:claim_run, scheduled_for)
    end
  end

  test "claim_run does not re-enqueue job when duplicate run already succeeded" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Duplicate Succeeded Run",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "status" => "ok" },
      notification_channels: %w[ email ]
    )
    scheduled_for = 1.minute.from_now.change(usec: 0)
    beeper.update_columns(next_run_at: scheduled_for, status: "active")
    beeper.runs.create!(scheduled_for: scheduled_for, status: :succeeded, signal_status: :ok)

    assert_no_enqueued_jobs only: RunBeeperJob do
      beeper.send(:claim_run, scheduled_for)
    end
  end

  test "poll_due_now records error signal and finishes firing when assigned runner is offline" do
    offline_runner = @account.runners.create!(name: "Offline-Node", status: "offline")
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Runner Offline Probe",
      cron: "*/5 * * * *",
      timezone: "UTC",
      runner: offline_runner,
      config: { "status" => "ok" },
      notification_channels: %w[ email ]
    )
    beeper.update_columns(next_run_at: 1.minute.ago, status: "active")

    assert_no_enqueued_jobs only: RunBeeperJob do
      Beeper.poll_due_now
    end

    beeper.reload
    assert beeper.active?
    assert_equal 1, beeper.runs.count
    run = beeper.runs.first
    assert_equal "failed", run.status
    assert_equal "error", run.signal_status
    assert_equal "Runner offline", run.signal_result["title"]
    assert_includes run.signal_result["message"], "Offline-Node"
    assert_equal 1, beeper.consecutive_failures
  end

  test "poll_due_now records error signal when no online runners match runner_tag" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Tag Offline Probe",
      cron: "*/5 * * * *",
      timezone: "UTC",
      runner_tag: "nonexistent-tag",
      config: { "status" => "ok" },
      notification_channels: %w[ email ]
    )
    beeper.update_columns(next_run_at: 1.minute.ago, status: "active")

    Beeper.poll_due_now

    beeper.reload
    assert beeper.active?
    assert_equal 1, beeper.runs.count
    run = beeper.runs.first
    assert_equal "failed", run.status
    assert_equal "error", run.signal_status
    assert_includes run.signal_result["message"], "nonexistent-tag"
  end

  test "reclaim_stale handles pending run when online runner becomes offline" do
    online_runner = @account.runners.create!(name: "Initially-Online")
    online_runner.update_columns(status: "online", last_seen_at: 5.seconds.ago)

    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Runner Stale Pending",
      cron: "*/5 * * * *",
      timezone: "UTC",
      runner: online_runner,
      config: { "status" => "ok" },
      notification_channels: %w[ email ]
    )
    beeper.update_columns(next_run_at: 1.minute.ago, status: "active")

    # Initial claim while online creates pending run
    Beeper.poll_due_now
    beeper.reload
    assert beeper.firing?
    assert_equal "pending", beeper.runs.last.status

    # Runner goes offline and beeper firing becomes stale
    online_runner.update_columns(status: "offline", last_seen_at: 70.seconds.ago)
    beeper.update_columns(updated_at: 3.minutes.ago)

    Beeper.reclaim_stale_firing

    beeper.reload
    assert beeper.active?
    run = beeper.runs.last
    assert_equal "failed", run.status
    assert_equal "error", run.signal_status
    assert_equal "Runner offline", run.signal_result["title"]
  end

  test "reclaim_stale handles running run that timed out without reporting results" do
    online_runner = @account.runners.create!(name: "Crash-Node")
    online_runner.update_columns(status: "online", last_seen_at: 5.seconds.ago)

    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Runner Crash Test",
      cron: "*/5 * * * *",
      timezone: "UTC",
      runner: online_runner,
      config: { "status" => "ok" },
      notification_channels: %w[ email ]
    )

    run = beeper.runs.create!(
      scheduled_for: 10.minutes.ago,
      status: "running",
      runner: online_runner,
      claimed_at: 10.minutes.ago,
      created_at: 10.minutes.ago,
      updated_at: 6.minutes.ago
    )
    beeper.update_columns(status: "firing", updated_at: 6.minutes.ago)

    Beeper.reclaim_stale_firing

    beeper.reload
    assert beeper.active?
    run.reload
    assert_equal "failed", run.status
    assert_equal "error", run.signal_status
    assert_equal "Runner execution timed out", run.signal_result["title"]
    assert_includes run.signal_result["message"], "Crash-Node"
  end

  private

  def echo_manifest
    {
      "manifest_version" => 1,
      "slug" => "echo",
      "name" => "Echo",
      "version" => "1.0.0",
      "author" => "Beep",
      "schedule" => { "default_cron" => "*/5 * * * *" }
    }
  end
end
