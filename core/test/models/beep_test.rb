require "test_helper"

class BeepTest < ActiveSupport::TestCase
  setup do
    @account = accounts(:john_account)
  end

  test "once beep copies run_at to next_run_at on create" do
    run_at = 1.hour.from_now.change(usec: 0)
    beep = Beep.create!(
      account: @account,
      kind: :once,
      message: "Call mom",
      run_at: run_at
    )

    assert_equal run_at, beep.next_run_at
    assert_equal "UTC", beep.timezone
    assert beep.active?
  end

  test "once beep requires message and run_at" do
    beep = Beep.new(account: @account, kind: :once)

    assert_not beep.valid?
    assert beep.errors[:message].any?
    assert beep.errors[:run_at].any?
  end

  test "once beep rejects cron" do
    beep = Beep.new(
      account: @account,
      kind: :once,
      message: "Call mom",
      run_at: 1.hour.from_now,
      cron: "0 9 * * *"
    )

    assert_not beep.valid?
    assert beep.errors[:cron].any?
  end

  test "once beep rejects a run_at in the past" do
    beep = Beep.new(
      account: @account,
      kind: :once,
      message: "Call mom",
      run_at: 1.hour.ago
    )

    assert_not beep.valid?
    assert beep.errors[:run_at].any?
  end

  test "once beep accepts a run_at in the future" do
    beep = Beep.new(
      account: @account,
      kind: :once,
      message: "Call mom",
      run_at: 1.hour.from_now
    )

    assert beep.valid?
  end
end
