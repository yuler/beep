class Channel::Delivery < ApplicationRecord
  DEFAULT_TTL = 30.minutes

  belongs_to :channel
  belongs_to :beep_run, class_name: "Beep::Run", optional: true

  enum :status, %w[ pending claimed succeeded failed expired ].index_by(&:itself), default: "pending"

  before_validation :assign_default_expires_at, on: :create

  validates :status, presence: true

  scope :due_for_cli, -> { pending.where(expires_at: Time.current..) }
  scope :stale_pending, -> { pending.where(expires_at: ...Time.current) }

  def expired?
    expires_at.present? && expires_at < Time.current
  end

  def claim!
    return false if expired?

    claimed = self.class.where(id: id, status: "pending").update_all(status: "claimed", claimed_at: Time.current, updated_at: Time.current) == 1
    if claimed
      self.status = "claimed"
      self.claimed_at = Time.current
    end
    claimed
  end

  def succeed!
    return false unless pending? || claimed?

    update!(status: :succeeded)
  end

  def fail!(error_message = nil)
    return false unless pending? || claimed?

    msg = error_message.to_s.strip.truncate(2048)
    merged = (payload || {}).merge("error" => msg.presence).compact
    update!(status: :failed, payload: merged)
  end

  private
    def assign_default_expires_at
      self.expires_at ||= DEFAULT_TTL.from_now
    end
end
