class ChannelAuthorization < ApplicationRecord
  DEFAULT_TTL = 15.minutes
  DEFAULT_INTERVAL = 5 # seconds
  USER_CODE_CHARSET = "BCDFGHJKMNPQRSTVWXYZ23456789".freeze

  belongs_to :account, optional: true
  belongs_to :user, optional: true
  belongs_to :channel, optional: true

  enum :status, %w[ pending approved access_denied expired ].index_by(&:itself), default: "pending"

  has_secure_token :device_code

  before_validation :generate_user_code, on: :create
  before_validation :set_expires_at, on: :create

  validates :device_code, presence: true, uniqueness: true
  validates :user_code, presence: true, uniqueness: true
  validates :expires_at, presence: true

  scope :active, -> { where(status: "pending").where("expires_at > ?", Time.current) }

  def self.create_request!(channel_name: nil)
    create!(channel_name: channel_name.presence)
  end

  def expired?
    status == "expired" || expires_at <= Time.current
  end

  def approve!(user:, name: nil)
    return false if expired? || status != "pending"

    target_name = name.presence || channel_name.presence || "CLI Channel"
    new_channel = user.account.channels.create!(
      user: user,
      kind: "cli",
      name: target_name
    )

    update!(
      account: user.account,
      user: user,
      channel: new_channel,
      channel_name: target_name,
      status: "approved"
    )
  end

  def deny!
    return false if expired?

    update!(status: "access_denied")
  end

  def poll!
    touch(:last_polled_at)
    if expired? && status == "pending"
      update!(status: "expired")
    end
  end

  private
    def generate_user_code
      return if user_code.present?

      loop do
        part1 = Array.new(4) { USER_CODE_CHARSET[SecureRandom.random_number(USER_CODE_CHARSET.length)] }.join
        part2 = Array.new(4) { USER_CODE_CHARSET[SecureRandom.random_number(USER_CODE_CHARSET.length)] }.join
        self.user_code = "#{part1}-#{part2}"
        break unless self.class.exists?(user_code: self.user_code)
      end
    end

    def set_expires_at
      self.expires_at ||= DEFAULT_TTL.from_now
    end
end
