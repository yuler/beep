class Runner::Authorization < ApplicationRecord
  DEFAULT_TTL = 15.minutes
  DEFAULT_INTERVAL = 5 # seconds
  USER_CODE_CHARSET = "BCDFGHJKMNPQRSTVWXYZ23456789".freeze

  belongs_to :account, optional: true
  belongs_to :user, optional: true
  belongs_to :runner, optional: true

  enum :status, %w[ pending approved consumed access_denied expired ].index_by(&:itself), default: "pending"

  has_secure_token :device_code

  normalizes :runner_name, with: ->(value) { value&.strip.presence }
  normalizes :tags, with: ->(value) {
    Array(value).map { |t| t.to_s.strip }.reject(&:blank?).uniq
  }

  before_validation :generate_user_code, on: :create
  before_validation :set_expires_at, on: :create

  validates :device_code, presence: true, uniqueness: true
  validates :user_code, presence: true, uniqueness: true
  validates :expires_at, presence: true

  scope :active, -> { where(status: "pending").where("expires_at > ?", Time.current) }

  def self.create_request!(runner_name: nil, tags: nil, metadata: nil)
    retries = 3
    begin
      create!(
        runner_name: runner_name.presence,
        tags: tags.presence || [],
        metadata: metadata.presence || {}
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

  def approve!(user:, name: nil, tags: nil)
    with_lock do
      return false if expired? || status != "pending"

      target_name = name.presence || runner_name.presence || "CLI Runner"
      target_tags = tags.nil? ? self.tags : tags

      new_runner = user.account.runners.create!(
        name: target_name,
        tags: target_tags
      )

      update!(
        account: user.account,
        user: user,
        runner: new_runner,
        runner_name: target_name,
        tags: target_tags,
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
    return nil if expired?
    return nil unless runner.present?

    consumed = self.class.where(id: id, status: "approved").update_all(status: "consumed", updated_at: Time.current) == 1
    if consumed
      self.status = "consumed"
      runner
    end
  end

  def poll_interval_exceeded?
    last_polled_at.present? && (Time.current - last_polled_at) < (DEFAULT_INTERVAL - 1)
  end

  def poll!
    touch(:last_polled_at)
    if expired? && status.in?(%w[ pending approved ])
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
