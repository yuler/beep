class Beep < ApplicationRecord
  self.table_name = "beeps"

  belongs_to :account

  enum :kind, %w[ once recurring ].index_by(&:itself)
  enum :status, %w[ active paused completed cancelled ].index_by(&:itself)

  validates :message, presence: true, length: { maximum: 500 }
  validates :timezone, presence: true
  validates :run_at, presence: true, if: :once?
  validates :run_at, absence: true, if: :recurring?
  validates :cron, presence: true, if: :recurring?
  validates :cron, absence: true, if: :once?
  validate :run_at_in_future, if: :once?

  before_validation :sync_next_run_at_for_once, on: :create

  scope :due, -> { active.where(next_run_at: ..Time.current) }

  private
    def sync_next_run_at_for_once
      if once? && run_at.present?
        self.next_run_at = run_at
      end
    end

    def run_at_in_future
      return if run_at.blank?
      return if run_at > Time.current

      errors.add(:run_at, "must be in the future")
    end
end
