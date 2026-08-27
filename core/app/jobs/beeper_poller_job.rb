class BeeperPollerJob < ApplicationJob
  queue_as :default

  def perform
    BeeperInstall.poll_due_now
  end
end
