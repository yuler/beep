class BeeperInstall < ApplicationRecord
  self.table_name = "beeper_installs"

  POLL_BATCH_SIZE = 100
  STALE_FIRING_AFTER = 2.minutes
  RUNNING_STALE_AFTER = 5.minutes
  EXPIRED_AFTER = 1.hour
  TITLE_MAX_LENGTH = 80

  belongs_to :account
  belongs_to :beeper
  has_many :runs, class_name: "BeeperRun", dependent: :destroy
  has_many :beeps, dependent: :nullify

  enum :status, %w[ active paused completed cancelled firing ].index_by(&:itself)
  enum :alert_state, %w[ ok alerting ].index_by(&:itself)

  store_accessor :signal_metadata, :ping_token, :last_ping_at

  normalizes :title, with: ->(value) { value.strip.presence }

  validates :title, presence: true, length: { maximum: TITLE_MAX_LENGTH }
  validates :cron, presence: true
  validates :timezone, presence: true
  validates :alert_state, presence: true
  validates :consecutive_failures, numericality: { greater_than_or_equal_to: 0 }

  validate :timezone_is_iana
  validate :validate_cron_expression
  validate :validate_notification_channels

  before_validation :assign_default_notification_channels
  before_validation :sync_next_run_at
  before_validation :ensure_ping_token_if_needed

  scope :due, -> { active.where(next_run_at: ..Time.current) }

  class << self
    def find_by_ping_token(token)
      return nil if token.blank?

      where("json_extract(signal_metadata, '$.ping_token') = ?", token).first
    end

    def poll_due_now
      due.order(:next_run_at).limit(POLL_BATCH_SIZE).each(&:claim_due)
      reclaim_stale_firing
    end

    def reclaim_stale_firing
      firing.where(updated_at: ..STALE_FIRING_AFTER.ago).find_each(&:reclaim_stale)
    end
  end

  def trigger_run!
    scheduled_for = Time.current
    update!(status: :firing, next_run_at: nil)

    run = begin
      runs.create!(scheduled_for: scheduled_for, status: :pending)
    rescue ActiveRecord::RecordNotUnique
      runs.find_by!(scheduled_for: scheduled_for)
    end
    run.deliver_later
    run
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

  def failure_threshold
    beeper&.failure_threshold || 2
  end

  def effective_config
    cfg = (config || {}).deep_stringify_keys
    if beeper&.webhook_ingest?
      cfg["last_ping_at"] = last_ping_at
      cfg["ping_token"] = ping_token
    end
    cfg
  end

  def record_ping
    meta = (signal_metadata || {}).merge("last_ping_at" => Time.current.utc.iso8601)
    update_columns(signal_metadata: meta, updated_at: Time.current)
  end

  def expired?(scheduled_for)
    return false if scheduled_for.nil?

    scheduled_for < EXPIRED_AFTER.ago
  end

  def notify_from!(signal)
    channels = Array(notification_channels).presence || account.owner_user.notification_channels
    account.beeps.create!(
      kind: :once,
      title: signal.title.presence || title,
      body: signal.message,
      timezone: timezone,
      notification_channels: channels,
      beeper_install: self
    )
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
    runs.find_by!(scheduled_for: scheduled_for)
  end

  def assign_default_notification_channels
    if Array(notification_channels).empty? && account&.owner_user
      self.notification_channels = account.owner_user.notification_channels
    end
  end

  def sync_next_run_at
    if (new_record? && next_run_at.nil?) || (persisted? && will_save_change_to_cron?)
      self.next_run_at = calculate_next_run_at
    end
  end

  def ensure_ping_token_if_needed
    return unless beeper&.webhook_ingest?

    self.ping_token ||= SecureRandom.alphanumeric(32)
  end

  def timezone_is_iana
    return if timezone.blank?

    errors.add(:timezone, "is invalid") unless IanaTimezone.valid?(timezone)
  end

  def validate_cron_expression
    return if cron.blank?

    tz = timezone.presence || IanaTimezone::DEFAULT
    if Fugit.parse("#{cron} #{tz}").nil?
      errors.add(:cron, "is invalid")
    end
  end

  def validate_notification_channels
    return if notification_channels.blank?

    invalid = Array(notification_channels) - User::NOTIFICATION_CHANNELS
    if invalid.any?
      errors.add(:notification_channels, "contains unsupported channels: #{invalid.join(', ')}")
    end
  end
end
