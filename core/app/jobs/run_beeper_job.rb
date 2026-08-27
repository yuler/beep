class RunBeeperJob < ApplicationJob
  queue_as :checks

  def perform(beeper_run)
    beeper_run.execute_now
  end
end
