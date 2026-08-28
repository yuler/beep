class BeeperPollerJob < ApplicationJob
  def perform
    Beeper.poll_due_now
  end
end
