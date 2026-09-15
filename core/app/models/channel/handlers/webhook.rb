# Webhook channels are not implemented yet. Creation is rejected via
# validate_config and delivery raises loudly instead of silently dropping
# notifications (which would mark the run as succeeded without sending).
class Channel::Handlers::Webhook < Channel::Handlers::Base
  class << self
    def deliver_beep(channel, beep, run: nil)
      raise NotImplementedError, "Webhook channel delivery is not implemented"
    end

    def deliver_test!(channel)
      raise NotImplementedError, "Webhook channel delivery is not implemented"
    end

    def validate_config(channel)
      channel.errors.add(:base, "Webhook channels are not supported yet")
    end
  end
end
