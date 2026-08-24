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
end
