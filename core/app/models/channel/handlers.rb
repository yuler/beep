module Channel::Handlers
  HANDLERS = {
    "cli" => "Channel::Handlers::Cli",
    "email" => "Channel::Handlers::Email",
    "web_push" => "Channel::Handlers::WebPush",
    "webhook" => "Channel::Handlers::Webhook"
  }.freeze

  def self.for(kind)
    handler_name = HANDLERS[kind.to_s]
    raise NotImplementedError, "Unhandled channel kind: #{kind}" unless handler_name

    handler_name.constantize
  end
end
