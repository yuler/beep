class PruneChannelAuthorizationsJob < ApplicationJob
  def perform
    Channel::Authorization.expire_pending_now
  end
end
