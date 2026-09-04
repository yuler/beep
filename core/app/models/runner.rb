class Runner < ApplicationRecord
  # rt = runner token
  TOKEN_PREFIX = "beep_rt_"
  OFFLINE_TIMEOUT = 60.seconds
  NAME_MAX_LENGTH = 80

  belongs_to :account
  has_many :jobs, class_name: "Runner::Job", dependent: :destroy
  has_many :runs, class_name: "Runner::Run", dependent: :destroy

  has_secure_token prefix: TOKEN_PREFIX
  self.filter_attributes += [ :token ]

  enum :status, %w[ online idle offline ].index_by(&:itself), default: "offline"

  normalizes :name, with: ->(value) { value&.strip.presence }

  validates :name, presence: true, length: { maximum: NAME_MAX_LENGTH }

  before_validation :normalize_tags

  scope :online_or_idle, -> { where(status: %w[ online idle ]).where(last_seen_at: OFFLINE_TIMEOUT.ago..) }

  class << self
    def find_by_raw_token(token)
      if token.present?
        find_by(token: token.to_s.strip)
      end
    end

    def mark_stale_offline!
      where(status: %w[ online idle ])
        .where(last_seen_at: ..OFFLINE_TIMEOUT.ago)
        .or(where(status: %w[ online idle ], last_seen_at: nil))
        .update_all(status: "offline")
    end
  end

  def touch_activity!(version: nil, os: nil, arch: nil, hostname: nil, ip_address: nil, status: "idle")
    attrs = {
      status: status.in?(%w[ online idle ]) ? status : "idle",
      last_seen_at: Time.current
    }
    attrs[:version] = version if version.present?
    attrs[:os] = os if os.present?
    attrs[:arch] = arch if arch.present?
    attrs[:hostname] = hostname if hostname.present?
    attrs[:ip_address] = ip_address if ip_address.present?

    update_columns(attrs)
  end

  def matches_tag?(tag)
    return true if tag.blank?

    Array(tags).map(&:to_s).include?(tag.to_s)
  end

  def online?
    status.in?(%w[ online idle ]) && last_seen_at.present? && last_seen_at >= OFFLINE_TIMEOUT.ago
  end

  def token_prefix
    token&.first(12)
  end

  def regenerate_token!
    regenerate_token
  end

  private

  def normalize_tags
    self.tags = Array(tags).map { |t| t.to_s.strip }.reject(&:blank?).uniq
  end
end
