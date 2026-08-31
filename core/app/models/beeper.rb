class Beeper < ApplicationRecord
  self.table_name = "beepers"

  POLL_BATCH_SIZE = 100
  STALE_FIRING_AFTER = 2.minutes
  RUNNING_STALE_AFTER = 5.minutes
  EXPIRED_AFTER = 1.hour
  TITLE_MAX_LENGTH = 80
  BODY_MAX_LENGTH = 2000

  belongs_to :account
  belongs_to :beeper_app
  has_many :runs, class_name: "BeeperRun", dependent: :destroy
  has_many :beeps, dependent: :nullify

  enum :status, %w[ active paused completed cancelled firing ].index_by(&:itself)
  enum :alert_state, %w[ ok pending alerting recovering ].index_by(&:itself)

  store_accessor :signal_metadata, :ping_token, :last_ping_at, :consecutive_recoveries

  normalizes :title, with: ->(value) { value.strip.presence }
  normalizes :body, with: ->(value) { value&.strip.presence }

  validates :title, presence: true, length: { maximum: TITLE_MAX_LENGTH }
  validates :body, length: { maximum: BODY_MAX_LENGTH }, allow_nil: true
  validates :cron, presence: true
  validates :timezone, presence: true
  validates :alert_state, presence: true
  validates :consecutive_failures, numericality: { greater_than_or_equal_to: 0 }

  validate :timezone_is_iana
  validate :validate_cron_expression
  validate :validate_notification_channels
  validate :validate_config_inputs
  validate :validate_beeper_app_has_receiver, on: :create

  before_validation :assign_default_notification_channels
  before_validation :sync_next_run_at
  before_validation :ensure_ping_token_if_needed

  scope :due, -> { active.where(next_run_at: ..Time.current) }

  class << self
    def find_by_ping_token(token, beeper_app_slug: nil)
      return nil if token.blank?

      scope = where("json_extract(signal_metadata, '$.ping_token') = ?", token)
      scope = scope.joins(:beeper_app).where(beeper_apps: { slug: beeper_app_slug }) if beeper_app_slug.present?
      scope.first
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
    update!(status: :firing, next_run_at: nil) if active?

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

  def alert_policy_name
    alert_policy_config["policy"] || "consecutive_failures"
  end

  def alert_policy_config
    # Precedence: Beeper config["alerting"] > BeeperApp manifest["alerting"] > defaults
    instance_alerting = (config || {})["alerting"]
    return instance_alerting if instance_alerting.is_a?(Hash) && instance_alerting.present?

    beeper_app&.alert_policy_config || {}
  end

  def failure_threshold
    (alert_policy_config["failure_threshold"] || beeper_app&.failure_threshold || 2).to_i
  end

  def recovery_threshold
    (alert_policy_config["recovery_threshold"] || beeper_app&.recovery_threshold || 1).to_i
  end

  def consecutive_recoveries
    (super || 0).to_i
  end

  def effective_config
    cfg = (config || {}).deep_stringify_keys
    if beeper_app&.webhook_ping?
      cfg["last_ping_at"] = last_ping_at
      cfg["ping_token"] = ping_token
      cfg["beeper_created_at"] = created_at&.utc&.iso8601
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
    title_text = (signal.title.presence || title).to_s.strip.truncate(Beep::TITLE_MAX_LENGTH)
    body_text = signal.message.to_s.strip.truncate(Beep::BODY_MAX_LENGTH)

    account.beeps.create!(
      kind: :once,
      title: title_text,
      body: body_text.presence,
      timezone: timezone,
      notification_channels: channels,
      beeper: self
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
    run = runs.find_by!(scheduled_for: scheduled_for)
    run.deliver_later if run.pending?
    run
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
    return unless beeper_app&.webhook_ping?

    self.ping_token ||= SecureRandom.alphanumeric(32)
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
      return
    end

    if beeper_app.present?
      min_interval = beeper_app.min_interval_seconds.to_i
      if min_interval > 0
        t1 = parsed.next_time(Time.current)
        t2 = t1 ? parsed.next_time(t1) : nil
        if t1 && t2 && (t2.to_i - t1.to_i) < min_interval
          errors.add(:cron, "interval cannot be shorter than #{min_interval} seconds")
        end
      end
    end
  end

  def validate_notification_channels
    return if notification_channels.blank?

    invalid = Array(notification_channels) - User::NOTIFICATION_CHANNELS
    if invalid.any?
      errors.add(:notification_channels, "contains unsupported channels: #{invalid.join(', ')}")
    end
  end

  def validate_beeper_app_has_receiver
    return if beeper_app.nil?
    return if beeper_app.receiver_class.present?

    errors.add(:beeper_app, "is not installable: no receiver implementation is available")
  end

  def validate_config_inputs
    return if beeper_app.nil?

    inputs = beeper_app.inputs
    return if inputs.blank?

    cfg = (config || {}).deep_stringify_keys

    inputs.each do |input|
      name = input["name"]
      value = cfg[name]
      required = input["required"] == true
      input_type = input["type"]
      label = input["label"] || name

      if required && (value.nil? || (value.is_a?(String) && value.strip.empty?))
        errors.add(:config, "#{label} is required")
        next
      end

      next if value.nil? || (value.is_a?(String) && value.strip.empty?)

      case input_type
      when "number"
        num = Float(value, exception: false)
        if num.nil?
          errors.add(:config, "#{label} must be a number")
        else
          min = input["min"]
          max = input["max"]
          if min && num < min
            errors.add(:config, "#{label} must be at least #{min}")
          end
          if max && num > max
            errors.add(:config, "#{label} must be at most #{max}")
          end
        end
      when "url"
        uri = URI.parse(value.to_s) rescue nil
        unless uri.is_a?(URI::HTTP) || uri.is_a?(URI::HTTPS)
          errors.add(:config, "#{label} must be a valid http or https URL")
        end
      when "enum"
        options = Array(input["options"]).map(&:to_s)
        if options.present? && !options.include?(value.to_s)
          errors.add(:config, "#{label} must be one of: #{options.join(', ')}")
        end
      end
    end
  end
end
