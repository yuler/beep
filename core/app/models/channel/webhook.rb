# TODO: implement webhook channel delivery and config validation
module Channel::Webhook
  class << self
    def deliver_beep(channel, beep, run: nil)
      # TODO: implement webhook delivery
    end

    def deliver_test!(channel)
      # TODO: implement webhook test delivery
    end

    def validate_config(channel)
      # TODO: implement webhook config validation (URL, headers, secret, etc.)
    end
  end
end
