class RunnerJob < ApplicationRecord
  self.table_name = "runner_jobs"

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
  has_many :runs, class_name: "RunnerRun", dependent: :destroy

  enum :status, %w[ active paused firing ].index_by(&:itself), default: "active"

  normalizes :name, with: ->(value) { value&.strip.presence }
  normalizes :slug, with: ->(value) { value&.strip&.downcase.presence }

  validates :name, presence: true, length: { maximum: NAME_MAX_LENGTH }
  validates :slug, presence: true, format: { with: SLUG_FORMAT }, uniqueness: { scope: :runner_id }
  validates :cron, presence: true
  validates :timezone, presence: true
  validates :timeout_seconds, numericality: { greater_than_or_equal_to: TIMEOUT_MIN, less_than_or_equal_to: TIMEOUT_MAX }
  validate :timezone_is_iana
  validate :validate_cron_expression
  validate :runner_belongs_to_account

  before_validation :sync_next_run_at
  before_validation :assign_account_from_runner

  scope :due, -> { active.where(next_run_at: ..Time.current) }

  class << self
    def poll_due_now
      Runner.mark_stale_offline!
      due.order(:next_run_at).limit(POLL_BATCH_SIZE).each(&:claim_due)
      reclaim_stale_firing
    end

    def reclaim_stale_firing
      firing.where(updated_at: ..STALE_FIRING_AFTER.ago).find_each(&:reclaim_stale)
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
      elsif !runner.online? || run.created_at < STALE_FIRING_AFTER.ago
        run.record_result!(
          status: :error,
          title: "Runner offline",
          message: "Assigned runner '#{runner.name}' is offline or did not claim the task",
          run_status: :failed
        )
      else
        touch
      end
    elsif run.running?
      if run.updated_at < RUNNING_STALE_AFTER.ago
        run.record_result!(
          status: :error,
          title: "Runner execution timed out",
          message: "Runner '#{runner.name}' claimed the task but did not report a result",
          run_status: :failed
        )
      end
    else
      finish_firing(last_run_at: run.scheduled_for)
    end
  end

  def finish_firing(last_run_at:)
    reload

    next_time = calculate_next_run_at(from: Time.current)
    if paused?
      update!(last_run_at: last_run_at, next_run_at: next_time)
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
    return false if scheduled_for.nil?

    scheduled_for < EXPIRED_AFTER.ago
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
    return if runner.blank? || account_id.blank?
    return if runner.account_id == account_id

    errors.add(:runner, "must belong to the same account")
  end

  def sync_next_run_at
    if (new_record? && next_run_at.nil?) || (persisted? && will_save_change_to_cron?)
      self.next_run_at = calculate_next_run_at
    end
  end

  def timezone_is_iana
    return if timezone.blank?

    errors.add(:timezone, "is invalid") unless IanaTimezone.valid?(timezone)
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
