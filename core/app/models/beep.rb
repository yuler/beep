class Beep < ApplicationRecord
  self.table_name = "beeps"

  POLL_BATCH_SIZE = 100
  STALE_FIRING_AFTER = 2.minutes
  RUNNING_STALE_AFTER = 5.minutes
  EXPIRED_AFTER = 1.hour
  TITLE_MAX_LENGTH = 80
  BODY_MAX_LENGTH = 2000

  belongs_to :account
  belongs_to :plugin, optional: true
  has_many :runs, class_name: "BeepRun", dependent: :destroy

  enum :kind, %w[ once recurring ].index_by(&:itself)
  enum :status, %w[ active paused completed cancelled firing ].index_by(&:itself)
  enum :alert_state, %w[ ok alerting ].index_by(&:itself)

  normalizes :title, with: ->(value) { value.strip.presence }
  normalizes :body, with: ->(value) { value&.strip.presence }

  validates :title, presence: true, length: { maximum: TITLE_MAX_LENGTH }
  validates :body, length: { maximum: BODY_MAX_LENGTH }, allow_nil: true
  validates :timezone, presence: true
  validates :run_at, absence: true, if: :recurring?
  validates :cron, presence: true, if: :recurring?
  validates :cron, absence: true, if: :once?
  validates :alert_state, presence: true
  validates :consecutive_failures, numericality: { greater_than_or_equal_to: 0 }

  validate :timezone_is_iana
  validate :validate_cron_expression, if: :recurring?
  validate :validate_plugin_configuration, if: :plugin_id?

  before_validation :sync_run_attributes
  before_validation :ensure_ping_token_if_needed
  after_create_commit :deliver_if_due_on_create

  scope :due, -> { active.where(next_run_at: ..Time.current) }
  scope :plugins, -> { where.not(plugin_id: nil) }
  scope :reminders, -> { where(plugin_id: nil) }

  class << self
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
    attributes = { status: :firing, next_run_at: nil }
    attributes[:run_at] = run_at || scheduled_for if once?
    update!(attributes)

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
      if recurring?
        update!(status: :active, next_run_at: calculate_next_run_at)
      else
        update!(status: :active)
      end
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

    next_run_at = once? ? nil : calculate_next_run_at(from: Time.current)
    if paused?
      attributes = { last_run_at: last_run_at }
      attributes[:next_run_at] = next_run_at if recurring?
      update!(attributes)
    elsif firing?
      if once?
        update!(status: :completed, next_run_at: nil, last_run_at: last_run_at)
      else
        update!(status: :active, next_run_at: next_run_at, last_run_at: last_run_at)
      end
    else
      update!(last_run_at: last_run_at)
    end
  end

  def calculate_next_run_at(from: Time.current)
    if recurring? && cron.present?
      tz = timezone.presence || IanaTimezone::DEFAULT
      parsed = Fugit.parse("#{cron} #{tz}")
      if parsed
        next_time = parsed.next_time(from)
        next_time ? Time.at(next_time.to_i).utc : nil
      end
    end
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

  def plugin?
    plugin_id.present?
  end

  def failure_threshold
    plugin&.failure_threshold || 2
  end

  def effective_plugin_config
    cfg = (plugin_config || {}).deep_stringify_keys
    if plugin&.webhook_ingest?
      cfg["last_ping_at"] = last_ping_at
      cfg["ping_token"] = ping_token
    end
    cfg
  end

  def record_ping
    update_columns(last_ping_at: Time.current, updated_at: Time.current)
  end

  def expired_plugin_run?(scheduled_for)
    return false unless plugin?

    scheduled_for < EXPIRED_AFTER.ago
  end

  def push_payload(run: nil)
    options = { data: { url: web_url, badge: 1 } }

    if run&.check_result.present?
      result_title = run.check_result["title"]
      result_message = run.check_result["message"]
      display_title = result_title.presence || title
      options[:body] = result_message if result_message.present?
      { title: display_title, options: options }
    else
      text = body_text
      options[:body] = text if text.present?
      { title: title, options: options }
    end
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

    def sync_run_attributes
      if once?
        if new_record?
          self.run_at ||= Time.current
          self.next_run_at = run_at
        elsif will_save_change_to_run_at?
          self.next_run_at = run_at
        end
      elsif recurring?
        if new_record? || will_save_change_to_cron? || will_save_change_to_timezone?
          self.next_run_at = calculate_next_run_at
        end
      end
    end

    def validate_cron_expression
      return if cron.blank?

      parsed = Fugit.parse(cron)
      if parsed.nil? || !parsed.is_a?(Fugit::Cron)
        errors.add(:cron, "is not a valid cron expression")
      end
    end

    def timezone_is_iana
      if timezone.present? && !IanaTimezone.valid?(timezone)
        errors.add(:timezone, "is invalid")
      end
    end

    def validate_plugin_configuration
      return unless plugin?

      unless recurring?
        errors.add(:kind, "must be recurring for plugin beeps")
      end
    end

    def ensure_ping_token_if_needed
      if plugin&.webhook_ingest? && ping_token.blank?
        self.ping_token = SecureRandom.hex(16)
      end
    end

    def deliver_if_due_on_create
      return unless once? && active? && next_run_at.present? && next_run_at <= Time.current

      claim_due
    end
end
