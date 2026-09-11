class Channel < ApplicationRecord
  TOKEN_PREFIX = "beep_ct_" # ct = channel token
  NAME_MAX_LENGTH = 80
  KINDS = %w[ cli email web_push webhook ].freeze
  PERMITTED_ENDPOINT_HOSTS = %w[
    jmt17.google.com
    fcm.googleapis.com
    updates.push.services.mozilla.com
    web.push.apple.com
    notify.windows.com
  ].freeze

  belongs_to :account, default: -> { user&.account }
  belongs_to :user
  has_many :deliveries, class_name: "ChannelDelivery", dependent: :destroy

  enum :kind, KINDS.index_by(&:itself), default: "cli"
  enum :status, %w[ active disabled ].index_by(&:itself), default: "active"

  has_secure_token prefix: TOKEN_PREFIX

  store_accessor :config, :endpoint, :p256dh_key, :auth_key, :user_agent

  normalizes :name, with: ->(value) { value&.strip.presence }

  validates :name, presence: true, length: { maximum: NAME_MAX_LENGTH }
  validates :kind, presence: true, inclusion: { in: KINDS }
  validate :user_belongs_to_account
  validate :validate_web_push_config, if: :web_push?

  scope :active_cli, -> { active.cli }
  scope :for_endpoint, ->(ep) { where("json_extract(config, '$.endpoint') = ?", ep) }

  def self.upsert_web_push_for!(user, attributes)
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

  def touch_last_seen
    touch(:last_seen_at)
  end

  def masked_token
    return if token.blank?

    "#{token.first(12)}••••"
  end

  def deliver_test!
    send_push(test_payload)
  end

  def deliver_beep(beep, run: nil)
    send_push(beep.push_payload(run: run))
  end

  def resolved_endpoint_ip
    return @resolved_endpoint_ip if defined?(@resolved_endpoint_ip)

    @resolved_endpoint_ip = SsrfProtection.resolve_public_ip(endpoint_uri&.host)
  end

  private
    def self.user_agent_device_name(user_agent)
      return "Browser" if user_agent.blank?
      return "curl" if user_agent.to_s.start_with?("curl")

      user_agent.to_s.truncate(NAME_MAX_LENGTH)
    end

    def user_belongs_to_account
      if user.present? && account.present? && user.account_id != account_id
        errors.add(:user, "must belong to the same account")
      end
    end

    def endpoint_uri
      @endpoint_uri ||= URI.parse(endpoint) if endpoint.present?
    rescue URI::InvalidURIError
      nil
    end

    def validate_web_push_config
      if endpoint.blank?
        errors.add(:endpoint, "can't be blank")
      elsif endpoint_uri.nil?
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
