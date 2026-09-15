class PruneRunnerAuthorizationsJob < ApplicationJob
  RETENTION_PERIOD = 7.days

  def perform
    Runner::Authorization.expire_pending_now
    Runner::Authorization.where(status: %w[ expired consumed access_denied ])
      .where("updated_at < ?", RETENTION_PERIOD.ago)
      .delete_all
  end
end
