require "test_helper"

class BeepTest < ActiveSupport::TestCase
  include ActiveJob::TestHelper

  setup do
    @account = accounts(:john_account)
  end

  test "recipient_users is the account owner" do
    beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Call mom",
      run_at: 1.hour.from_now
    )

    assert_equal [ users(:john) ], beep.recipient_users
  end

  test "once beep copies run_at to next_run_at on create" do
    run_at = 1.hour.from_now.change(usec: 0)
    beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Call mom",
      run_at: run_at
    )

    assert_equal run_at, beep.next_run_at
    assert_equal "UTC", beep.timezone
    assert beep.active?
  end

  test "once beep requires title" do
    beep = Beep.new(account: @account, kind: :once)

    assert_not beep.valid?
    assert beep.errors[:title].any?
  end

  test "rejects an invalid timezone" do
    beep = Beep.new(
      account: @account,
      kind: :once,
      title: "Call mom",
      timezone: "Not/A_Zone"
    )

    assert_not beep.valid?
    assert beep.errors[:timezone].any?
  end

  test "accepts a valid IANA timezone" do
    beep = Beep.new(
      account: @account,
      kind: :once,
      title: "Call mom",
      timezone: "Asia/Shanghai"
    )

    assert beep.valid?
  end

  test "once beep defaults run_at and next_run_at to current time on create when omitted" do
    beep = Beep.create!(account: @account, kind: :once, title: "Call mom")

    assert_not_nil beep.run_at
    assert_equal beep.run_at, beep.next_run_at
  end

  test "once beep allows a blank body" do
    beep = Beep.new(
      account: @account,
      kind: :once,
      title: "Call mom",
      body: "   ",
      run_at: 1.hour.from_now
    )

    assert beep.valid?
    beep.save!
    assert_nil beep.body
  end

  test "once beep rejects a title longer than 80 characters" do
    beep = Beep.new(
      account: @account,
      kind: :once,
      title: "a" * 81,
      run_at: 1.hour.from_now
    )

    assert_not beep.valid?
    assert beep.errors[:title].any?
  end

  test "once beep rejects a body longer than 2000 characters" do
    beep = Beep.new(
      account: @account,
      kind: :once,
      title: "Call mom",
      body: "a" * 2001,
      run_at: 1.hour.from_now
    )

    assert_not beep.valid?
    assert beep.errors[:body].any?
  end

  test "body_text strips markdown from body" do
    beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Call mom",
      body: "Bring **milk** and [eggs](https://example.com)",
      run_at: 1.hour.from_now
    )

    assert_equal "Bring milk and eggs", beep.body_text
  end

  test "body_text is empty when the beep has no body" do
    beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Call mom",
      run_at: 1.hour.from_now
    )

    assert_equal "", beep.body_text
  end

  test "push payload uses title and body_text" do
    beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Call mom",
      body: "Bring **milk** and [eggs](https://example.com)",
      run_at: 1.hour.from_now
    )

    payload = beep.push_payload

    assert_equal "Call mom", payload[:title]
    assert_equal beep.body_text, payload.dig(:options, :body)
    assert_equal beep.web_url, payload.dig(:options, :data, :url)
  end

  test "push payload omits body when the beep has none" do
    beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Call mom",
      run_at: 1.hour.from_now
    )

    assert_nil beep.push_payload.dig(:options, :body)
  end

  test "once beep rejects cron" do
    beep = Beep.new(
      account: @account,
      kind: :once,
      title: "Call mom",
      run_at: 1.hour.from_now,
      cron: "0 9 * * *"
    )

    assert_not beep.valid?
    assert beep.errors[:cron].any?
  end

  test "once beep created with immediate/past run_at triggers delivery automatically" do
    assert_enqueued_with(job: DeliverBeepRunJob) do
      beep = Beep.create!(
        account: @account,
        kind: :once,
        title: "Immediate Beep"
      )

      assert beep.firing?
      run = beep.runs.sole
      assert run.pending?
    end
  end

  test "trigger_run! sets status to firing and creates a pending run" do
    beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Immediate Beep",
      run_at: 1.hour.from_now
    )

    run = beep.trigger_run!
    beep.reload

    assert beep.firing?
    assert_nil beep.next_run_at
    assert run.pending?
    assert_equal beep.id, run.beep_id
  end

  test "trigger_run! supports recurring beeps without validation errors" do
    beep = Beep.create!(
      account: @account,
      kind: :recurring,
      title: "Standup",
      cron: "0 9 * * *"
    )

    run = beep.trigger_run!
    beep.reload

    assert beep.firing?
    assert_nil beep.next_run_at
    assert_nil beep.run_at
    assert run.pending?

    beep.finish_firing(last_run_at: run.scheduled_for)
    assert beep.reload.active?
    assert_not_nil beep.next_run_at
  end

  test "recurring beep validates cron expression format" do
    beep = Beep.new(
      account: @account,
      kind: :recurring,
      title: "Invalid Cron",
      cron: "invalid cron syntax"
    )

    assert_not beep.valid?
    assert beep.errors[:cron].any?
  end

  test "recurring beep automatically sets next_run_at on creation" do
    beep = Beep.create!(
      account: @account,
      kind: :recurring,
      title: "Daily morning",
      cron: "0 9 * * *"
    )

    assert_not_nil beep.next_run_at
    assert beep.next_run_at > Time.current
  end

  test "pause! and resume! updates recurring beep status and next_run_at" do
    beep = Beep.create!(
      account: @account,
      kind: :recurring,
      title: "Daily morning",
      cron: "0 9 * * *"
    )

    beep.pause!
    assert beep.paused?

    beep.resume!
    assert beep.active?
    assert_not_nil beep.next_run_at
    assert beep.next_run_at > Time.current
  end

  test "finish_firing keeps a paused recurring beep paused" do
    beep = Beep.create!(
      account: @account,
      kind: :recurring,
      title: "Standup",
      cron: "0 9 * * *"
    )
    run = beep.trigger_run!
    beep.pause!

    Beep.find(beep.id).finish_firing(last_run_at: run.scheduled_for)

    beep.reload
    assert beep.paused?
    assert_equal run.scheduled_for.to_i, beep.last_run_at.to_i
  end

  test "finish_firing skips missed recurring slots after a delayed run" do
    travel_to Time.utc(2026, 8, 25, 10, 5, 0) do
      beep = Beep.create!(
        account: @account,
        kind: :recurring,
        title: "Every minute",
        timezone: "UTC",
        cron: "* * * * *"
      )
      beep.update_columns(status: "firing")
      beep.finish_firing(last_run_at: 1.hour.ago)

      assert beep.reload.active?
      assert beep.next_run_at > Time.current
    end
  end

  test "recurring beep next_run_at is the next 09:00 in Asia/Shanghai" do
    travel_to Time.utc(2026, 8, 25, 2, 0, 0) do
      beep = Beep.create!(
        account: @account,
        kind: :recurring,
        title: "Daily morning",
        timezone: "Asia/Shanghai",
        cron: "0 9 * * *"
      )

      assert_equal Time.utc(2026, 8, 26, 1, 0, 0), beep.next_run_at
    end
  end

  test "pause! and resume! reject illegal status transitions" do
    beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Call mom",
      run_at: 1.hour.from_now
    )
    beep.update_columns(status: "completed", next_run_at: nil)

    assert_raises ActiveRecord::RecordInvalid do
      beep.pause!
    end
    assert beep.reload.completed?

    assert_raises ActiveRecord::RecordInvalid do
      beep.resume!
    end
    assert beep.reload.completed?
  end
end
