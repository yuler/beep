class Push::Subscription < Channel
  default_scope { where(kind: :web_push) }

  def initialize(attributes = nil)
    super
    self.kind = "web_push"
    self.name ||= "Browser"
  end

  def self.upsert_for!(user, attributes)
    Channel.upsert_web_push_for!(user, attributes)
  end
end
