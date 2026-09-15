class PruneCliAuthorizationsJob < ApplicationJob
  def perform
    Cli::Authorization.expire_pending_now
  end
end
