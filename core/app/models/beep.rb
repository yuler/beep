class Beep < ApplicationRecord
  self.table_name = "beeps"

  POLL_BATCH_SIZE = 100
  STALE_FIRING_AFTER = 2.minutes
  RUNNING_STALE_AFTER = 5.minutes
  EXPIRED_AFTER = 1.hour
  TITLE_MAX_LENGTH = 80
  BODY_MAX_LENGTH = 2000
  TIMEZONE = "Asia/Shanghai"

  belongs_to :account
  has_many :runs, class_name: "BeepRun", dependent: :destroy

  enum :kind, %w[ once recurring ].index_by(&:itself)
  enum :status, %w[ active paused completed cancelled firing ].index_by(&:itself)

  normalizes :title, with: ->(value) { value.strip.presence }
  normalizes :body, with: ->(value) { value&.strip.presence }

  validates :title, presence: true, length: { maximum: TITLE_MAX_LENGTH }
  validates :body, length: { maximum: BODY_MAX_LENGTH }, allow_nil: true
  validates :timezone, presence: true
  validates :run_at, absence: true, if: :recurring?
  validates :cron, presence: true, if: :recurring?
  validates :cron, absence: true, if: :once?

  validate :validate_cron_expression, if: :recurring?

  before_validation :sync_run_attributes
  after_create_commit :deliver_if_due_on_create

  scope :due, -> { active.where(next_run_at: ..Time.current) }

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
    update!(status: :paused)
  end

  def resume!
    if recurring?
      update!(status: :active, next_run_at: calculate_next_run_at)
    else
      update!(status: :active)
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
    if once?
      update!(status: :completed, next_run_at: nil, last_run_at: last_run_at)
    else
      update!(status: :active, next_run_at: calculate_next_run_at(from: last_run_at || Time.current), last_run_at: last_run_at)
    end
  end

  def calculate_next_run_at(from: Time.current)
    return nil unless recurring? && cron.present?

    tz = timezone.presence || TIMEZONE
    parsed = Fugit.parse("#{cron} #{tz}") || Fugit.parse(cron)
    return nil unless parsed

    next_time = parsed.next_time(from)
    next_time ? Time.at(next_time.to_i).utc : nil
  rescue StandardError
    nil
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

    def deliver_if_due_on_create
      return unless once? && active? && next_run_at.present? && next_run_at <= Time.current

      claim_due
    end
end
