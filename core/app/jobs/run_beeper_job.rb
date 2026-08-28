class RunBeeperJob < ApplicationJob
  def perform(beeper_run)
    beeper_run.execute_now
  end
end
