require "test_helper"

class BeepPollerTest < ActiveSupport::TestCase
  include ActiveJob::TestHelper

  setup do
    @account = accounts(:john_account)
  end

  test "poll claims a due once beep, inserts a run, and enqueues deliver" do
    beep = due_once_beep
    scheduled_for = beep.next_run_at

    assert_enqueued_with(job: DeliverBeepRunJob) do
      Beep.poll_due_now
    end

    beep.reload
    assert beep.firing?
    run = beep.runs.sole
    assert run.pending?
    assert_equal scheduled_for.to_i, run.scheduled_for.to_i
  end

  test "poll ignores recurring beeps even when next_run_at is due" do
    beep = Beep.create!(
      account: @account,
      kind: :recurring,
      message: "Standup",
      cron: "0 9 * * *"
    )
    beep.update_columns(next_run_at: 1.minute.ago)

    assert_no_enqueued_jobs only: DeliverBeepRunJob do
      Beep.poll_due_now
    end

    assert_equal 0, beep.runs.count
    assert beep.reload.active?
  end

  test "poll ignores paused once beeps" do
    beep = due_once_beep
    beep.update!(status: :paused)

    assert_no_enqueued_jobs only: DeliverBeepRunJob do
      Beep.poll_due_now
    end

    assert_equal 0, beep.runs.count
  end

  test "poll flags an expired once beep as expired and does not enqueue delivery" do
    beep = due_once_beep
    beep.update_columns(next_run_at: (Beep::EXPIRED_AFTER + 1.minute).ago)

    assert_no_enqueued_jobs only: DeliverBeepRunJob do
      Beep.poll_due_now
    end

    run = beep.runs.sole
    assert run.expired?
    assert beep.reload.completed?
    assert_nil beep.next_run_at
    assert_equal run.scheduled_for.to_i, beep.last_run_at.to_i
  end

  test "poll expires a stale firing beep whose pending run is too old to deliver" do
    beep = due_once_beep
    Beep.poll_due_now
    run = beep.runs.sole
    run.update_columns(scheduled_for: (Beep::EXPIRED_AFTER + 1.minute).ago)
    beep.update_columns(updated_at: 3.minutes.ago)
    clear_enqueued_jobs

    assert_no_enqueued_jobs only: DeliverBeepRunJob do
      Beep.poll_due_now
    end

    assert run.reload.expired?
    assert beep.reload.completed?
    assert_nil beep.next_run_at
  end

  test "poll is idempotent for the same due beep" do
    due_once_beep

    Beep.poll_due_now
    assert_no_difference -> { BeepRun.count } do
      Beep.poll_due_now
    end
  end

  test "poll re-enqueues a stale firing beep whose run is still pending" do
    beep = due_once_beep
    Beep.poll_due_now
    run = beep.runs.sole
    beep.update_columns(updated_at: 3.minutes.ago)
    clear_enqueued_jobs

    assert_enqueued_with(job: DeliverBeepRunJob, args: [ run ]) do
      Beep.poll_due_now
    end
  end

  test "poll does not reclaim a stale firing beep whose run is running" do
    beep = due_once_beep
    Beep.poll_due_now
    beep.runs.sole.update!(status: :running)
    beep.update_columns(updated_at: 3.minutes.ago)
    clear_enqueued_jobs

    assert_no_enqueued_jobs only: DeliverBeepRunJob do
      Beep.poll_due_now
    end
  end

  test "poll fails a stale firing beep whose run has been running too long" do
    beep = due_once_beep
    Beep.poll_due_now
    run = beep.runs.sole
    run.update_columns(status: "running", updated_at: (Beep::RUNNING_STALE_AFTER + 1.minute).ago)
    beep.update_columns(updated_at: 3.minutes.ago)

    Beep.poll_due_now

    assert run.reload.failed?
    assert beep.reload.completed?
    assert_nil beep.next_run_at
  end

  test "poll completes a stale firing beep whose run already succeeded" do
    beep = due_once_beep
    Beep.poll_due_now
    beep.runs.sole.update!(status: :succeeded)
    beep.update_columns(updated_at: 3.minutes.ago)

    Beep.poll_due_now

    assert beep.reload.completed?
    assert_nil beep.next_run_at
  end

  private
    def due_once_beep
      beep = Beep.create!(
        account: @account,
        kind: :once,
        message: "Call mom",
        run_at: 1.hour.from_now.change(usec: 0)
      )
      beep.update_columns(next_run_at: 1.minute.ago.change(usec: 0))
      beep
    end
end
