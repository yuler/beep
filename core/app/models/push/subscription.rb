class Push::Subscription < ApplicationRecord
  PERMITTED_ENDPOINT_HOSTS = %w[
    jmt17.google.com
    fcm.googleapis.com
    updates.push.services.mozilla.com
    web.push.apple.com
    notify.windows.com
  ].freeze

  belongs_to :account, default: -> { user.account }
  belongs_to :user

  validates :endpoint, presence: true
  validate :validate_endpoint_url

  # `user.push_subscriptions.create!` / `update!` wrap `#save!` in a SQLite
  # IMMEDIATE transaction. Use `new` + `save!` so the write lock is only held
  # for the INSERT/UPDATE itself.
  def self.upsert_for!(user, attributes)
    attributes = attributes.to_h.symbolize_keys.merge(user: user, account: user.account)
    subscription = find_by(user_id: user.id, endpoint: attributes[:endpoint])

    if subscription
      subscription.assign_attributes(attributes)
    else
      subscription = new(attributes)
    end

    subscription.save!
    subscription
  rescue ActiveRecord::RecordNotUnique
    subscription = find_by!(user_id: user.id, endpoint: attributes[:endpoint])
    subscription.assign_attributes(attributes)
    subscription.save!
    subscription
  end

  def deliver_test!
    send_push(test_payload)
  end

  def deliver_beep(beep)
    send_push(beep.push_payload)
  end

  def resolved_endpoint_ip
    return @resolved_endpoint_ip if defined?(@resolved_endpoint_ip)

    @resolved_endpoint_ip = SsrfProtection.resolve_public_ip(endpoint_uri&.host)
  end

  private
    def send_push(payload)
      WebPush.payload_send(
        message: payload.to_json,
        endpoint: endpoint,
        p256dh: p256dh_key,
        auth: auth_key,
        vapid: {
          subject: Rails.application.config.x.vapid.subject,
          public_key: Rails.application.config.x.vapid.public_key,
          private_key: Rails.application.config.x.vapid.private_key
        },
        urgency: "high"
      )
    end

    def endpoint_uri
      @endpoint_uri ||= URI.parse(endpoint) if endpoint.present?
    rescue URI::InvalidURIError
      nil
    end

    def validate_endpoint_url
      if endpoint_uri.nil?
        errors.add(:endpoint, "is not a valid URL")
      elsif endpoint_uri.scheme != "https"
        errors.add(:endpoint, "must use HTTPS")
      elsif !permitted_endpoint_host?
        errors.add(:endpoint, "is not a permitted push service")
      end
    end

    def permitted_endpoint_host?
      host = endpoint_uri&.host&.downcase
      PERMITTED_ENDPOINT_HOSTS.any? { |permitted| host&.end_with?(permitted) }
    end

    def test_payload
      {
        title: "Beep",
        options: {
          body: "This is a test notification.",
          tag: "beep-test",
          renotify: true,
          data: { url: "/#{account.slug}/settings", badge: 1 }
        }
      }
    end
end
