class Channel < ApplicationRecord
  TOKEN_PREFIX = "beep_ct_" # ct = channel token
  NAME_MAX_LENGTH = 80
  KINDS = %w[ device email web_push webhook ].freeze

  belongs_to :account
  belongs_to :user
  has_many :deliveries, class_name: "ChannelDelivery", dependent: :destroy

  enum :kind, KINDS.index_by(&:itself), default: "device"
  enum :status, %w[ active disabled ].index_by(&:itself), default: "active"

  has_secure_token prefix: TOKEN_PREFIX

  normalizes :name, with: ->(value) { value&.strip.presence }

  validates :name, presence: true, length: { maximum: NAME_MAX_LENGTH }
  validates :kind, presence: true, inclusion: { in: KINDS }
  validate :user_belongs_to_account

  scope :active_devices, -> { active.device }

  def touch_last_seen
    touch(:last_seen_at)
  end

  def masked_token
    return if token.blank?

    "#{token.first(12)}••••"
  end

  private
    def user_belongs_to_account
      if user.present? && account.present? && user.account_id != account_id
        errors.add(:user, "must belong to the same account")
      end
    end
end
