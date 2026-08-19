class User < ApplicationRecord
  NOTIFICATION_CHANNELS = %w[ email web_push ].freeze
  DEFAULT_NOTIFICATION_CHANNELS = %w[ email web_push ].freeze

  include Role

  belongs_to :account
  belongs_to :identity, optional: true

  has_many :push_subscriptions, class_name: "Push::Subscription", dependent: :delete_all

  before_validation :assign_default_notification_channels, on: :create
  before_validation :normalize_notification_channels
  validates :name, presence: true
  validate :notification_channels_are_allowed

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
end
