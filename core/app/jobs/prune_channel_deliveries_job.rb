class PruneChannelDeliveriesJob < ApplicationJob
  RETENTION_PERIOD = 30.days

  def perform
    Channel::Delivery.expire_stale!
    Channel::Delivery.where(status: %w[ succeeded failed expired ])
      .where("updated_at < ?", RETENTION_PERIOD.ago)
      .delete_all
  end
end
