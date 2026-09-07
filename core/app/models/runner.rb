class Runner < ApplicationRecord
  TOKEN_PREFIX = "beep_rt_" # rt = runner token
  OFFLINE_TIMEOUT = 60.seconds
  NAME_MAX_LENGTH = 80

  belongs_to :account
  has_many :jobs, dependent: :destroy
  has_many :runs, dependent: :destroy

  enum :status, %w[ offline online idle ].index_by(&:itself), default: "offline"

  has_secure_token prefix: TOKEN_PREFIX

  normalizes :name, with: ->(value) { value&.strip.presence }
  normalizes :tags, with: ->(value) {
    Array(value).map { |t| t.to_s.strip }.reject(&:blank?).uniq
  }

  validates :name, presence: true, length: { maximum: NAME_MAX_LENGTH }

  class << self
    def mark_stale_offline
      where(status: %w[ online idle ])
        .where(last_seen_at: ..OFFLINE_TIMEOUT.ago)
        .or(where(status: %w[ online idle ], last_seen_at: nil))
        .update_all(status: "offline")
    end
  end

  def touch_activity(version: nil, os: nil, arch: nil, hostname: nil, ip_address: nil, status: "idle")
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
end
