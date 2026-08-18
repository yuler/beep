require "test_helper"

class DeliverBeepRunJobTest < ActiveSupport::TestCase
  test "perform delivers the run and completes the beep" do
    account = accounts(:john_account)
    beep = Beep.create!(
      account: account,
      kind: :once,
      message: "Call mom",
      run_at: 1.hour.from_now.change(usec: 0)
    )
    beep.update_columns(next_run_at: 1.minute.ago)
    Beep.poll_due_now
    run = beep.runs.sole

    DeliverBeepRunJob.perform_now(run)

    assert run.reload.succeeded?
    assert beep.reload.completed?
  end
end
