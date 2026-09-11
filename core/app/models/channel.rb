class Channel < ApplicationRecord
  TOKEN_PREFIX = "beep_ct_" # ct = channel token
  NAME_MAX_LENGTH = 80
  KINDS = %w[ cli email web_push webhook ].freeze
  PERMITTED_ENDPOINT_HOSTS = Channel::WebPush::PERMITTED_ENDPOINT_HOSTS

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
  validate :validate_kind_config

  scope :active_cli, -> { active.cli }
  scope :for_endpoint, ->(ep) { where("json_extract(config, '$.endpoint') = ?", ep) }

  def self.upsert_web_push_for!(user, attributes)
    Channel::WebPush.upsert_for!(user, attributes)
  end

  ONLINE_TIMEOUT = 5.minutes

  def touch_last_seen
    return if last_seen_at && last_seen_at > 30.seconds.ago

    touch(:last_seen_at)
  end

  def online?
    active? && last_seen_at.present? && last_seen_at >= ONLINE_TIMEOUT.ago
  end

  def masked_token
    return if token.blank?

    "#{token.first(12)}••••"
  end

  def deliver_beep(beep, run: nil)
    handler.deliver_beep(self, beep, run: run)
  end

  def deliver_test!
    handler.deliver_test!(self)
  end

  def handler
    case kind
    when "web_push" then Channel::WebPush
    when "cli"      then Channel::Cli
    when "email"    then Channel::Email
    when "webhook"  then Channel::Webhook
    else
      raise NotImplementedError, "Unhandled channel kind: #{kind}"
    end
  end

  def resolved_endpoint_ip
    Channel::WebPush.resolved_endpoint_ip(self)
  end

  private
    def user_belongs_to_account
      if user.present? && account.present? && user.account_id != account_id
        errors.add(:user, "must belong to the same account")
      end
    end

    def validate_kind_config
      handler.validate_config(self) if respond_to?(:handler) && handler.respond_to?(:validate_config)
    end
end
