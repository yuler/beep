Rails.application.configure do
  config.x.vapid.private_key = ENV["VAPID_PRIVATE_KEY"].presence
  config.x.vapid.public_key  = ENV["VAPID_PUBLIC_KEY"].presence

  # VAPID `sub` is a contact URI push services use to reach the app operator.
  # Derive it from EMAIL_SENDER (the single configured contact address); fall
  # back to the web-push gem default so the claim is never nil.
  config.x.vapid.subject     = "mailto:#{ENV["EMAIL_SENDER"].presence || "sender@example.com"}"
end
