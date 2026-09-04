class Runner < ApplicationRecord
  self.table_name = "runners"

  TOKEN_PREFIX = "beep_rt_"
  OFFLINE_TIMEOUT = 60.seconds
  NAME_MAX_LENGTH = 80

  belongs_to :account
  has_many :jobs, class_name: "RunnerJob", dependent: :destroy
  has_many :runs, class_name: "RunnerRun", dependent: :destroy

  enum :status, %w[ online idle offline ].index_by(&:itself), default: "offline"

  attr_reader :raw_token

  normalizes :name, with: ->(value) { value&.strip.presence }

  validates :name, presence: true, length: { maximum: NAME_MAX_LENGTH }
  validates :token_digest, presence: true, uniqueness: true
  validates :token_prefix, presence: true

  before_validation :generate_token_if_needed, on: :create
  before_validation :normalize_tags

  scope :online_or_idle, -> { where(status: %w[ online idle ]).where(last_seen_at: OFFLINE_TIMEOUT.ago..) }

  class << self
    def find_by_raw_token(token)
      return nil if token.blank?

      digest = Digest::SHA256.hexdigest(token.to_s.strip)
      find_by(token_digest: digest)
    end

    def mark_stale_offline!
      where(status: %w[ online idle ])
        .where(last_seen_at: ..OFFLINE_TIMEOUT.ago)
        .or(where(status: %w[ online idle ], last_seen_at: nil))
        .update_all(status: "offline")
    end
  end

  def touch_activity!(version: nil, os: nil, arch: nil, hostname: nil, ip_address: nil, allow_exec: nil, status: "idle")
    attrs = {
      status: status.in?(%w[ online idle ]) ? status : "idle",
      last_seen_at: Time.current
    }
    attrs[:version] = version if version.present?
    attrs[:os] = os if os.present?
    attrs[:arch] = arch if arch.present?
    attrs[:hostname] = hostname if hostname.present?
    attrs[:ip_address] = ip_address if ip_address.present?
    attrs[:allow_exec] = allow_exec unless allow_exec.nil?

    update_columns(attrs)
  end

  def matches_tag?(tag)
    return true if tag.blank?

    Array(tags).map(&:to_s).include?(tag.to_s)
  end

  def online?
    status.in?(%w[ online idle ]) && last_seen_at.present? && last_seen_at >= OFFLINE_TIMEOUT.ago
  end

  def regenerate_token!
    generate_token
    save!
  end

  private

  def generate_token_if_needed
    generate_token if token_digest.blank?
  end

  def generate_token
    token = "#{TOKEN_PREFIX}#{SecureRandom.hex(24)}"
    self.token_digest = Digest::SHA256.hexdigest(token)
    self.token_prefix = token[0, 12]
    @raw_token = token
  end

  def normalize_tags
    self.tags = Array(tags).map { |t| t.to_s.strip }.reject(&:blank?).uniq
  end
end
