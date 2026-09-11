class ChannelDelivery < ApplicationRecord
  DEFAULT_TTL = 30.minutes

  belongs_to :channel
  belongs_to :beep_run, optional: true

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

    update(status: :claimed, claimed_at: Time.current)
  end

  def succeed!
    update!(status: :succeeded)
  end

  def fail!(error_message = nil)
    update!(status: :failed)
  end

  private
    def assign_default_expires_at
      self.expires_at ||= DEFAULT_TTL.from_now
    end
end
