class User < ApplicationRecord
  NOTIFICATION_CHANNELS = %w[ email web_push ].freeze
  DEFAULT_NOTIFICATION_CHANNELS = %w[ email ].freeze

  include Role

  enum :timezone_source, %w[ detected manual ].index_by(&:itself)

  belongs_to :account
  belongs_to :identity, optional: true

  has_many :push_subscriptions, class_name: "Push::Subscription", dependent: :delete_all

  before_validation :assign_default_notification_channels, on: :create
  before_validation :normalize_notification_channels
  before_validation :normalize_timezone
  validates :name, presence: true
  validate :notification_channels_are_allowed
  validate :timezone_is_iana
  validate :timezone_source_is_allowed

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

  def assign_timezone(name:, source:)
    timezone_name = name.to_s.strip.presence
    timezone_source_name = source.to_s.strip.presence

    if timezone_source_name == "detected"
      if timezone.blank? && timezone_name
        self.timezone = timezone_name
        self.timezone_source = "detected"
      end
    elsif timezone_name
      self.timezone = timezone_name
      self.timezone_source = "manual"
    end
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

    def normalize_timezone
      self.timezone = timezone.to_s.strip.presence
      self.timezone_source = timezone_source.to_s.strip.presence
      if timezone.blank?
        self.timezone_source = nil
      end
    end

    def timezone_is_iana
      if timezone.present? && !IanaTimezone.valid?(timezone)
        errors.add(:timezone, "is invalid")
      end
    end

    def timezone_source_is_allowed
      if timezone.present? && self.class.timezone_sources.exclude?(timezone_source)
        errors.add(:timezone_source, "is invalid")
      end
    end
end
