require "test_helper"

class BeepPollerJobTest < ActiveSupport::TestCase
  include ActiveJob::TestHelper

  test "perform claims due once beeps" do
    account = accounts(:john_account)
    beep = Beep.create!(
      account: account,
      kind: :once,
      message: "Call mom",
      run_at: 1.hour.from_now.change(usec: 0)
    )
    beep.update_columns(next_run_at: 1.minute.ago)

    BeepPollerJob.perform_now

    assert beep.reload.firing?
    assert_equal 1, beep.runs.count
  end
end
