require "test_helper"

class BeepRunTest < ActiveSupport::TestCase
  test "rejects a second run for the same beep and scheduled_for" do
    account = accounts(:john_account)
    beep = Beep.create!(
      account: account,
      kind: :once,
      title: "Call mom",
      run_at: 1.hour.from_now.change(usec: 0)
    )
    scheduled_for = beep.next_run_at
    beep.runs.create!(scheduled_for: scheduled_for, status: :pending)

    duplicate = beep.runs.new(scheduled_for: scheduled_for, status: :pending)
    assert_raises(ActiveRecord::RecordNotUnique) { duplicate.save(validate: false) }
  end
end
