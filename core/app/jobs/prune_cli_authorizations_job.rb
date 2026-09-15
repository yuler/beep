class PruneCliAuthorizationsJob < ApplicationJob
  RETENTION_PERIOD = 7.days

  def perform
    Cli::Authorization.expire_pending_now
    Cli::Authorization.where(status: %w[ expired consumed access_denied ])
      .where("updated_at < ?", RETENTION_PERIOD.ago)
      .delete_all
  end
end
