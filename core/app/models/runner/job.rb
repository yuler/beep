class Runner::Job < ApplicationRecord
  POLL_BATCH_SIZE = 100
  STALE_FIRING_AFTER = 2.minutes
  RUNNING_STALE_AFTER = 5.minutes
  EXPIRED_AFTER = 1.hour
  NAME_MAX_LENGTH = 80
  SLUG_FORMAT = /\A[a-z0-9][a-z0-9_-]*\z/
  TIMEOUT_MIN = 5
  TIMEOUT_MAX = 300

  belongs_to :account
  belongs_to :runner
  has_many :runs, class_name: "Runner::Run", foreign_key: :runner_job_id, inverse_of: :runner_job, dependent: :destroy

  enum :status, %w[ active paused firing ].index_by(&:itself), default: "active"

  normalizes :name, with: ->(value) { value&.strip.presence }
  normalizes :slug, with: ->(value) { value&.strip&.downcase.presence }

  before_validation :sync_next_run_at
  before_validation :assign_account_from_runner

  validates :name, presence: true, length: { maximum: NAME_MAX_LENGTH }
  validates :slug, presence: true, format: { with: SLUG_FORMAT }, uniqueness: { scope: :runner_id }
  validates :cron, presence: true
  validates :timezone, presence: true
  validates :timeout_seconds, numericality: { greater_than_or_equal_to: TIMEOUT_MIN, less_than_or_equal_to: TIMEOUT_MAX }
  validate :timezone_is_iana
  validate :validate_cron_expression
  validate :runner_belongs_to_account

  scope :due, -> { active.where(next_run_at: ..Time.current) }

  class << self
    def poll_due_now
      Runner.mark_stale_offline
      due.order(:next_run_at).limit(POLL_BATCH_SIZE).each(&:claim_due)
      reclaim_stale_firing
    end

    def reclaim_stale_firing
      firing.where(updated_at: ..STALE_FIRING_AFTER.ago).find_each(&:reclaim_stale)

      # pause! leaves open pending/running runs outside the firing set — still reclaim them.
      paused.joins(:runs).where(runner_runs: { status: %w[ pending running ] }).distinct.find_each(&:reclaim_stale)
    end
  end

  def trigger_run!
    scheduled_for = Time.current
    update!(status: :firing, next_run_at: nil) if active?

    begin
      runs.create!(scheduled_for: scheduled_for, status: :pending, runner: runner)
    rescue ActiveRecord::RecordNotUnique
      runs.find_by!(scheduled_for: scheduled_for)
    end
  end

  def pause!
    if active? || firing?
      update!(status: :paused)
    else
      errors.add(:status, "cannot be paused")
      raise ActiveRecord::RecordInvalid, self
    end
  end

  def resume!
    if paused?
      update!(status: :active, next_run_at: calculate_next_run_at)
    else
      errors.add(:status, "cannot be resumed")
      raise ActiveRecord::RecordInvalid, self
    end
  end

  def claim_due
    scheduled_for = next_run_at
    claimed = self.class.where(id: id, status: :active).update_all(
      status: "firing",
      updated_at: Time.current
    )
    if claimed == 1
      self.status = "firing"
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
      elsif paused?
        run.record_result!(
          status: :error,
          title: "Job paused",
          message: "Job '#{name}' was paused before the pending run was claimed",
          run_status: :failed,
          from_statuses: %w[ pending ]
        )
      elsif !runner.online?
        run.record_result!(
          status: :error,
          title: "Runner offline",
          message: "Assigned runner '#{runner.name}' is offline or unreachable",
          run_status: :failed,
          from_statuses: %w[ pending ]
        )
      else
        touch
      end
    elsif run.running?
      stale_threshold = (timeout_seconds || 60).seconds + 30.seconds
      started_at = run.claimed_at || run.created_at
      if started_at < stale_threshold.ago
        run.record_result!(
          status: :error,
          title: "Runner execution timed out",
          message: "Runner '#{runner.name}' claimed the task but did not report a result within #{timeout_seconds || 60}s",
          run_status: :failed,
          from_statuses: %w[ running ]
        )
      else
        touch
      end
    else
      finish_firing(last_run_at: run.scheduled_for)
    end
  end

  def finish_firing(last_run_at:)
    reload

    next_time = calculate_next_run_at(from: Time.current)
    if paused?
      update!(last_run_at: last_run_at)
    elsif firing?
      update!(status: :active, next_run_at: next_time, last_run_at: last_run_at)
    else
      update!(last_run_at: last_run_at)
    end
  end

  def calculate_next_run_at(from: Time.current)
    if cron.present?
      tz = timezone.presence || IanaTimezone::DEFAULT
      parsed = Fugit.parse("#{cron} #{tz}")
      if parsed
        next_time = parsed.next_time(from)
        next_time ? Time.at(next_time.to_i).utc : nil
      end
    end
  end

  def expired?(scheduled_for)
    scheduled_for.present? && scheduled_for < EXPIRED_AFTER.ago
  end

  private
    def claim_run(scheduled_for)
      if expired?(scheduled_for)
        runs.create!(scheduled_for: scheduled_for, status: :expired, runner: runner)
        finish_firing(last_run_at: scheduled_for)
      elsif !runner.online?
        run = runs.create!(scheduled_for: scheduled_for, status: :pending, runner: runner)
        run.record_result!(
          status: :error,
          title: "Runner offline",
          message: "Assigned runner '#{runner.name}' is offline or unreachable",
          run_status: :failed
        )
        run
      else
        run = runs.create!(scheduled_for: scheduled_for, status: :pending, runner: runner)
        touch
        run
      end
    rescue ActiveRecord::RecordNotUnique
      runs.find_by!(scheduled_for: scheduled_for)
    end

    def assign_account_from_runner
      self.account_id = runner.account_id if runner.present? && account_id.blank?
    end

    def runner_belongs_to_account
      if runner.present? && account_id.present? && runner.account_id != account_id
        errors.add(:runner, "must belong to the same account")
      end
    end

    def sync_next_run_at
      if (new_record? && next_run_at.nil?) || (persisted? && (will_save_change_to_cron? || will_save_change_to_timezone?))
        self.next_run_at = calculate_next_run_at
      end
    end

    def timezone_is_iana
      if timezone.present? && !IanaTimezone.valid?(timezone)
        errors.add(:timezone, "is invalid")
      end
    end

    def validate_cron_expression
      return if cron.blank?

      tz = timezone.presence || IanaTimezone::DEFAULT
      parsed = Fugit.parse("#{cron} #{tz}")
      if parsed.nil?
        errors.add(:cron, "is invalid")
      end
    end
end
