class PruneBeeperRunsJob < ApplicationJob
  def perform
    BeeperRun.prune_expired_now
  end
end
