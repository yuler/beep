class Beep::Preview
  attr_reader :beep, :errors, :schedule_key, :schedule_display, :timezone, :kind, :run_at, :next_run_at, :cron

  def self.build(account:, user: nil, params:)
    new(account: account, user: user, params: params).build
  end

  def initialize(account:, user: nil, params:)
    @account = account
    @user = user
    @params = params
    @errors = []
  end

  def build
    @timezone = IanaTimezone.resolve(@user&.timezone, @params[:timezone])
    run_at_val = resolve_schedule_run_at(timezone: @timezone)
    @kind = @params[:kind].presence || (@params[:cron].present? ? "recurring" : "once")

    attrs = beep_attributes.merge(kind: @kind, timezone: @timezone)
    attrs[:run_at] = run_at_val if run_at_val.present?

    @beep = @account.beeps.new(attrs)
    @beep.validate
    @errors.concat(@beep.errors.full_messages)

    @run_at = @beep.run_at
    @next_run_at = @beep.next_run_at
    @cron = @beep.cron
    @schedule_key, @schedule_display = format_schedule(run_at: run_at_val, timezone: @timezone)

    self
  end

  def valid?
    @errors.empty?
  end

  def title
    @beep&.title.presence || @params[:title].to_s.strip
  end

  def body
    @beep&.body.presence || @params[:body].to_s.strip.presence
  end

  def intent
    @beep&.intent.presence || @params[:intent].to_s.strip.presence
  end

  def metadata
    @beep&.metadata || @params[:metadata]
  end

  # Prefer explicit params / already-set attrs. Do not call
  # assign_default_notification_channels on this dry-run object.
  def notification_channels
    if @params[:notification_channels].present?
      Array(@params[:notification_channels])
    elsif @beep&.notification_channels.present?
      @beep.notification_channels
    else
      []
    end
  end

  private
    def resolve_schedule_run_at(timezone:)
      if @params[:cron].present? && (@params[:in].present? || @params[:at].present?)
        @errors << "Cannot specify :cron together with :in or :at"
        return nil
      end

      Beep::ScheduleParser.resolve_run_at(
        in_val: @params[:in],
        at_val: @params[:at],
        run_at_val: @params[:run_at],
        timezone: timezone
      )
    rescue Beep::ScheduleParser::Error => e
      @errors << e.message
      nil
    end

    def beep_attributes
      raw = if @params.respond_to?(:permit)
        @params.permit(:title, :body, :run_at, :cron, :kind, :intent, notification_channels: [], metadata: {})
      else
        @params
      end
      raw.to_h.symbolize_keys.slice(:title, :body, :run_at, :cron, :kind, :intent, :notification_channels, :metadata)
    end

    def format_schedule(run_at:, timezone:)
      tz = Time.find_zone(timezone) || Time.find_zone("UTC")
      local_now = Time.current.in_time_zone(tz)

      if @kind == "recurring"
        cron_str = @cron.to_s.strip
        display = if @next_run_at.present?
          "#{cron_str} (Next: #{format_local_occurrence(@next_run_at, local_now, tz, timezone)})"
        else
          cron_str
        end
        [ "Cron", display ]
      elsif @params[:in].present?
        in_val = @params[:in].to_s.strip
        display = if run_at.present?
          "#{in_val} → #{format_local_occurrence(run_at, local_now, tz, timezone, with_seconds: true)}"
        else
          in_val
        end
        [ "Delay", display ]
      elsif @params[:at].present? || @params[:run_at].present? || run_at.present?
        at_val = (@params[:at].presence || @params[:run_at].presence || run_at&.in_time_zone(tz)&.strftime("%Y-%m-%d %H:%M")).to_s.strip
        display = if run_at.present?
          local_target = run_at.in_time_zone(tz)
          occurrence = format_local_occurrence(run_at, local_now, tz, timezone)
          if local_target.to_date == local_now.to_date || local_target.to_date == (local_now.to_date + 1.day)
            "#{at_val} (#{occurrence})"
          else
            diff_days = (local_target.to_date - local_now.to_date).to_i
            "#{at_val} (#{occurrence} · in #{diff_days} days)"
          end
        else
          at_val
        end
        [ "Run At", display ]
      else
        [ "Schedule", "instant (fires immediately)" ]
      end
    end

    # Shared Today/Tomorrow / absolute local formatting for the four CLI shapes.
    def format_local_occurrence(time, local_now, tz, timezone, with_seconds: false)
      local = time.in_time_zone(tz)
      clock = with_seconds ? local.strftime("%H:%M:%S") : local.strftime("%H:%M")
      stamp = "#{local.strftime('%Y-%m-%d %H:%M:%S')} #{timezone}"

      if local.to_date == local_now.to_date
        "Today at #{clock} · #{stamp}"
      elsif local.to_date == (local_now.to_date + 1.day)
        "Tomorrow at #{clock} · #{stamp}"
      else
        stamp
      end
    end
end
