class PruneBeeperRunsJob < ApplicationJob
  RETENTION = 30.days

  def perform
    BeeperRun.where(scheduled_for: ..RETENTION.ago).in_batches(of: 1_000) do |batch|
      batch.delete_all
    end
  end
end
