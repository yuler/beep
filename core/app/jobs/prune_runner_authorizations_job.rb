class PruneRunnerAuthorizationsJob < ApplicationJob
  def perform
    Runner::Authorization.expire_pending_now
  end
end
