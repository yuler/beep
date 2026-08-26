class RunCheckJob < ApplicationJob
  queue_as :checks

  def perform(beep_run)
    beep_run.execute_check_now
  end
end
