class Cli::Authorization < ApplicationRecord
  DEFAULT_TTL = 15.minutes
  DEFAULT_INTERVAL = 5 # seconds
  USER_CODE_CHARSET = "BCDFGHJKMNPQRSTVWXYZ23456789".freeze

  belongs_to :identity, optional: true
  belongs_to :access_token, class_name: "Identity::AccessToken", optional: true

  enum :status, %w[ pending approved consumed access_denied expired ].index_by(&:itself), default: "pending"

  has_secure_token :device_code

  normalizes :client_name, with: ->(value) { value&.strip.presence }

  before_validation :generate_user_code, on: :create
  before_validation :set_expires_at, on: :create

  validates :device_code, presence: true, uniqueness: true
  validates :user_code, presence: true, uniqueness: true
  validates :expires_at, presence: true

  scope :active, -> { where(status: "pending").where("expires_at > ?", Time.current) }

  def self.create_request!(client_name: nil)
    retries = 3
    begin
      create!(
        client_name: client_name.presence || "Beep CLI"
      )
    rescue ActiveRecord::RecordNotUnique
      retries -= 1
      retries >= 0 ? retry : raise
    end
  end

  def self.expire_pending_now
    where(status: "pending").where(expires_at: ..Time.current)
      .update_all(status: "expired", updated_at: Time.current)
  end

  def expired?
    status == "expired" || expires_at <= Time.current
  end

  def approve!(identity:, client_name: nil)
    with_lock do
      return false if expired? || status != "pending"

      target_name = client_name.presence || self.client_name.presence || "Beep CLI"
      new_token = identity.access_tokens.create!(
        description: target_name,
        permission: "write"
      )

      update!(
        identity: identity,
        access_token: new_token,
        client_name: target_name,
        status: "approved"
      )
    end
  rescue ActiveRecord::RecordInvalid => error
    errors.add(:base, error.message)
    false
  end

  def deny!
    return false if expired?

    denied = self.class.where(id: id, status: "pending").update_all(status: "access_denied", updated_at: Time.current) == 1
    self.status = "access_denied" if denied
    denied
  end

  def consume_token!
    return nil unless access_token.present?

    consumed = self.class.where(id: id, status: "approved").update_all(status: "consumed", updated_at: Time.current) == 1
    if consumed
      self.status = "consumed"
      access_token
    end
  end

  def poll_interval_exceeded?
    last_polled_at.present? && (Time.current - last_polled_at) < (DEFAULT_INTERVAL - 1)
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
