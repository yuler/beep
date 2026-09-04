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

  test "reclaim_stale handles running run that timed out" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Hung Probe",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "status" => "ok" },
      notification_channels: %w[ email ]
    )

    run = beeper.runs.create!(
      scheduled_for: 10.minutes.ago,
      status: "running",
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
    assert_equal "Probe execution timed out", run.signal_result["title"]
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
