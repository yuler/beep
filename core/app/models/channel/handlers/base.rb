class Channel::Handlers::Base
  class << self
    def deliver_beep(channel, beep, run: nil)
      raise NotImplementedError, "#{name}.deliver_beep is not implemented"
    end

    def deliver_test!(channel)
      raise NotImplementedError, "#{name}.deliver_test! is not implemented"
    end

    def validate_config(channel)
    end
  end
end
