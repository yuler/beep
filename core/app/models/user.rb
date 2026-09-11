class User < ApplicationRecord
  NOTIFICATION_CHANNELS = %w[ email web_push cli ].freeze
  DEFAULT_NOTIFICATION_CHANNELS = %w[ email ].freeze

  include Role

  belongs_to :account
  belongs_to :identity, optional: true

  has_many :push_subscriptions, class_name: "Push::Subscription", dependent: :delete_all
  has_many :channels, dependent: :destroy
  has_many :channel_deliveries, through: :channels, source: :deliveries

  enum :timezone_source, %w[ detected manual ].index_by(&:itself)

  normalizes :timezone, with: ->(value) { value&.strip.presence }

  before_validation :assign_default_notification_channels, on: :create
  before_validation :normalize_notification_channels
  before_validation :clear_timezone_source, unless: :timezone?
  after_create :create_default_email_channel

  validates :name, presence: true
  validates :timezone_source, presence: true, if: :timezone?
  validate :notification_channels_are_allowed
  validate :timezone_is_iana, if: :timezone?

  # TODO: deactivate user
  def deactivate
    transaction do
      # accesses.destroy_all
      # update! active: false, identity: nil
      # close_remote_connections
    end
  end

  def verified?
    verified_at.present?
  end

  def verify
    update!(verified_at: Time.current) unless verified?
  end

  def notification_channel?(name)
    Array(notification_channels).include?(name.to_s)
  end

  def email_channel_unsubscribe_token
    signed_id(purpose: :email_channel_unsubscribe, expires_in: 1.year)
  end

  def assign_timezone(name:, source: :manual)
    return if name.blank?
    return if source.to_s == "detected" && timezone?

    self.timezone = name
    self.timezone_source = source.to_s == "detected" ? :detected : :manual
  end

  private
    def assign_default_notification_channels
      if notification_channels.nil?
        self.notification_channels = DEFAULT_NOTIFICATION_CHANNELS.dup
      end
    end

    def normalize_notification_channels
      if notification_channels
        self.notification_channels = Array(notification_channels).map { |value| value.to_s.strip }.reject(&:blank?).uniq
      end
    end

    def notification_channels_are_allowed
      unknown = Array(notification_channels) - NOTIFICATION_CHANNELS
      if unknown.any?
        errors.add(:notification_channels, "is invalid")
      end
    end

    def clear_timezone_source
      self.timezone_source = nil
    end

    def timezone_is_iana
      errors.add(:timezone, "is invalid") unless IanaTimezone.valid?(timezone)
    end

    def create_default_email_channel
      return if channels.where(kind: :email).exists?

      email_name = identity&.email.presence || name.presence || "email"
      channels.create!(
        account: account,
        kind: :email,
        name: email_name
      )
    rescue ActiveRecord::RecordInvalid
      # ignore if already exists or invalid
    end
end
