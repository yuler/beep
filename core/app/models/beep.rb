class Beep < ApplicationRecord
  self.table_name = "beeps"

  POLL_BATCH_SIZE = 100
  STALE_FIRING_AFTER = 2.minutes

  belongs_to :account
  has_many :runs, class_name: "BeepRun", dependent: :destroy

  enum :kind, %w[ once recurring ].index_by(&:itself)
  enum :status, %w[ active paused completed cancelled firing ].index_by(&:itself)

  validates :message, presence: true, length: { maximum: 500 }
  validates :timezone, presence: true
  validates :run_at, presence: true, if: :once?
  validates :run_at, absence: true, if: :recurring?
  validates :cron, presence: true, if: :recurring?
  validates :cron, absence: true, if: :once?
  validate :run_at_in_future, if: :once?

  before_validation :sync_next_run_at_for_once, on: :create

  scope :due, -> { once.active.where(next_run_at: ..Time.current) }

  class << self
    def poll_due_now
      due.order(:next_run_at).limit(POLL_BATCH_SIZE).each(&:claim_due)
      reclaim_stale_firing
    end

    def reclaim_stale_firing
      firing.where(updated_at: ..STALE_FIRING_AFTER.ago).find_each(&:reclaim_stale)
    end
  end

  def claim_due
    scheduled_for = next_run_at
    claimed = self.class.where(id: id, status: :active).update_all(
      status: "firing",
      updated_at: Time.current
    )
    if claimed == 1
      enqueue_run(scheduled_for)
    end
  end

  def reclaim_stale
    run = runs.order(:created_at).last
    if run.nil?
      enqueue_run(next_run_at)
    elsif run.pending?
      touch
      run.deliver_later
    elsif !run.running?
      finish_firing(last_run_at: run.scheduled_for)
    end
  end

  def finish_firing(last_run_at:)
    update!(status: :completed, next_run_at: nil, last_run_at: last_run_at)
  end

  def web_url
    "#{Rails.application.config.x.web_origin}/#{account.slug}/beeps/#{id}"
  end

  def push_payload
    {
      title: "Beep",
      options: {
        body: message,
        data: { url: web_url, badge: 1 }
      }
    }
  end

  private
    def enqueue_run(scheduled_for)
      run = runs.create!(scheduled_for: scheduled_for, status: :pending)
      touch
      run.deliver_later
    rescue ActiveRecord::RecordNotUnique
      runs.find_by!(scheduled_for: scheduled_for).deliver_later
    end

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
