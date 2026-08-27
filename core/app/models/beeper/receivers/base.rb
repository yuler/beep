class Beeper::Receivers::Base
  attr_reader :config

  def initialize(config: {})
    @config = (config || {}).deep_stringify_keys
  end

  def self.call(config:)
    new(config: config).call
  end

  def call
    raise NotImplementedError, "#{self.class.name}#call must be implemented"
  end
end
