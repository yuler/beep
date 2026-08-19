class Beep < ApplicationRecord
  self.table_name = "beeps"

  POLL_BATCH_SIZE = 100
  STALE_FIRING_AFTER = 2.minutes
  RUNNING_STALE_AFTER = 5.minutes
  EXPIRED_AFTER = 1.hour
  TITLE_MAX_LENGTH = 80
  BODY_MAX_LENGTH = 2000

  belongs_to :account
  has_many :runs, class_name: "BeepRun", dependent: :destroy

  enum :kind, %w[ once recurring ].index_by(&:itself)
  enum :status, %w[ active paused completed cancelled firing ].index_by(&:itself)

  normalizes :title, with: ->(value) { value.strip.presence }
  normalizes :body, with: ->(value) { value&.strip.presence }

  validates :title, presence: true, length: { maximum: TITLE_MAX_LENGTH }
  validates :body, length: { maximum: BODY_MAX_LENGTH }, allow_nil: true
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
      claim_run(scheduled_for)
    end
  end

  def reclaim_stale
    run = runs.order(:created_at).last
    if run.nil?
      claim_run(next_run_at)
    elsif run.pending?
      if expired?(run.scheduled_for)
        run.update!(status: :expired)
        finish_firing(last_run_at: run.scheduled_for)
      else
        touch
        run.deliver_later
      end
    elsif run.running?
      if run.updated_at < RUNNING_STALE_AFTER.ago
        run.update!(status: :failed)
        finish_firing(last_run_at: run.scheduled_for)
      end
    else
      finish_firing(last_run_at: run.scheduled_for)
    end
  end

  def finish_firing(last_run_at:)
    update!(status: :completed, next_run_at: nil, last_run_at: last_run_at)
  end

  def web_url
    "#{Rails.application.config.x.web_origin}/#{account.slug}/beeps/#{id}"
  end

  def recipient_users
    [ account.owner_user ]
  end

  def body_text
    Beep::Plaintext.from_markdown(body)
  end

  def push_payload
    options = { data: { url: web_url, badge: 1 } }
    text = body_text
    if text.present?
      options[:body] = text
    end

    { title: title, options: options }
  end

  private
    def claim_run(scheduled_for)
      if expired?(scheduled_for)
        runs.create!(scheduled_for: scheduled_for, status: :expired)
        finish_firing(last_run_at: scheduled_for)
      else
        run = runs.create!(scheduled_for: scheduled_for, status: :pending)
        touch
        run.deliver_later
      end
    rescue ActiveRecord::RecordNotUnique
      run = runs.find_by(scheduled_for: scheduled_for)
      finish_firing(last_run_at: run.scheduled_for) if run&.expired?
    end

    def expired?(scheduled_for)
      scheduled_for < EXPIRED_AFTER.ago
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
