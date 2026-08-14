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

  def self.upsert_for!(user, attributes)
    attributes = attributes.to_h.symbolize_keys
    subscription = user.push_subscriptions.find_by(endpoint: attributes[:endpoint])

    if subscription
      subscription.update!(attributes)
      subscription
    else
      user.push_subscriptions.create!(attributes)
    end
  rescue ActiveRecord::RecordNotUnique
    subscription = user.push_subscriptions.find_by!(endpoint: attributes[:endpoint])
    subscription.update!(attributes)
    subscription
  end

  def resolved_endpoint_ip
    return @resolved_endpoint_ip if defined?(@resolved_endpoint_ip)

    @resolved_endpoint_ip = SsrfProtection.resolve_public_ip(endpoint_uri&.host)
  end

  private
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
      elsif resolved_endpoint_ip.nil?
        errors.add(:endpoint, "resolves to a private or invalid IP address")
      end
    end

    def permitted_endpoint_host?
      host = endpoint_uri&.host&.downcase
      PERMITTED_ENDPOINT_HOSTS.any? { |permitted| host&.end_with?(permitted) }
    end
end
