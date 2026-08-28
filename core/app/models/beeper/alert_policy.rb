class Beeper::AlertPolicy
  Decision = Data.define(
    :should_notify,
    :next_alert_state,
    :next_consecutive_failures,
    :next_consecutive_recoveries,
    :is_recovery
  )

  DEFAULT_POLICY = "consecutive_failures".freeze

  class << self
    def for(beeper)
      policy_name = beeper.alert_policy_name.presence || DEFAULT_POLICY
      policy_class = policy_registry[policy_name.to_s] || ConsecutiveFailures
      policy_class.new(beeper)
    end

    def policy_registry
      @policy_registry ||= {
        "consecutive_failures" => ConsecutiveFailures,
        "windowed" => Windowed
      }
    end

    def register(name, klass)
      policy_registry[name.to_s] = klass
    end
  end

  attr_reader :beeper

  def initialize(beeper)
    @beeper = beeper
  end

  def evaluate(signal:)
    raise NotImplementedError, "#{self.class.name} must implement #evaluate(signal:)"
  end

  protected

  def policy_config
    beeper.alert_policy_config
  end
end
