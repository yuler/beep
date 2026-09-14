class Channel < ApplicationRecord
  TOKEN_PREFIX = "beep_ct_" # ct = channel token
  NAME_MAX_LENGTH = 80
  KINDS = %w[ cli email web_push webhook ].freeze
  Cli = Handlers::Cli
  Email = Handlers::Email
  WebPush = Handlers::WebPush
  Webhook = Handlers::Webhook

  PERMITTED_ENDPOINT_HOSTS = Handlers::WebPush::PERMITTED_ENDPOINT_HOSTS

  belongs_to :account, default: -> { user&.account }
  belongs_to :user
  has_many :deliveries, class_name: "Channel::Delivery", dependent: :destroy
  has_many :authorizations, class_name: "Channel::Authorization", dependent: :destroy

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
    Handlers::WebPush.upsert_for!(user, attributes)
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

    "#{TOKEN_PREFIX}••••"
  end

  def deliver_beep(beep, run: nil)
    handler.deliver_beep(self, beep, run: run)
  end

  def deliver_test!
    handler.deliver_test!(self)
  end

  def handler
    Handlers.for(kind)
  end

  def resolved_endpoint_ip
    Handlers::WebPush.resolved_endpoint_ip(self)
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
