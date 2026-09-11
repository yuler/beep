module Channel::WebPush
  PERMITTED_ENDPOINT_HOSTS = %w[
    jmt17.google.com
    fcm.googleapis.com
    updates.push.services.mozilla.com
    web.push.apple.com
    notify.windows.com
  ].freeze

  module_function

  def deliver_beep(channel, beep, run: nil)
    send_push(channel, beep.push_payload(run: run))
  end

  def deliver_test!(channel)
    send_push(channel, test_payload(channel))
  end

  def send_push(channel, payload)
    WebPush.payload_send(
      message: payload.to_json,
      endpoint: channel.endpoint,
      p256dh: channel.p256dh_key,
      auth: channel.auth_key,
      vapid: {
        subject: Rails.application.config.x.vapid.subject,
        public_key: Rails.application.config.x.vapid.public_key,
        private_key: Rails.application.config.x.vapid.private_key
      },
      urgency: "high"
    )
  end

  def upsert_for!(user, attributes)
    attrs = attributes.to_h.symbolize_keys
    endpoint_val = attrs[:endpoint]
    channel = user.channels.where(kind: :web_push).find { |c| c.endpoint == endpoint_val }

    name_val = attrs[:name].presence || user_agent_device_name(attrs[:user_agent])
    config_data = {
      endpoint: attrs[:endpoint],
      p256dh_key: attrs[:p256dh_key],
      auth_key: attrs[:auth_key],
      user_agent: attrs[:user_agent]
    }.compact.stringify_keys

    if channel
      channel.assign_attributes(name: name_val, config: channel.config.merge(config_data))
    else
      channel = user.channels.new(
        account: user.account,
        kind: :web_push,
        name: name_val,
        config: config_data
      )
    end

    channel.save!
    channel
  end

  def user_agent_device_name(user_agent)
    return "Browser" if user_agent.blank?
    return "curl" if user_agent.to_s.start_with?("curl")

    user_agent.to_s.truncate(Channel::NAME_MAX_LENGTH)
  end

  def resolved_endpoint_ip(channel)
    uri = endpoint_uri(channel)
    SsrfProtection.resolve_public_ip(uri&.host)
  end

  def endpoint_uri(channel)
    URI.parse(channel.endpoint) if channel.endpoint.present?
  rescue URI::InvalidURIError
    nil
  end

  def validate_config(channel)
    if channel.endpoint.blank?
      channel.errors.add(:endpoint, "can't be blank")
    elsif (uri = endpoint_uri(channel)).nil?
      channel.errors.add(:endpoint, "is not a valid URL")
    elsif uri.scheme != "https"
      channel.errors.add(:endpoint, "must use HTTPS")
    elsif !permitted_endpoint_host?(uri)
      channel.errors.add(:endpoint, "is not a permitted push service")
    end
  end

  def permitted_endpoint_host?(uri)
    host = uri&.host&.downcase
    PERMITTED_ENDPOINT_HOSTS.any? { |permitted| host&.end_with?(permitted) }
  end

  def test_payload(channel)
    {
      title: "Beep",
      options: {
        body: "This is a test notification.",
        tag: "beep-test",
        renotify: true,
        data: { url: "/#{channel.account.slug}/settings", badge: 1 }
      }
    }
  end
end
