class BeeperPollerJob < ApplicationJob
  queue_as :default

  def perform
    Beeper.poll_due_now
  end
end
