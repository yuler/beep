class Beep::Preview
  attr_reader :beep, :errors, :schedule_key, :schedule_display, :timezone, :kind, :run_at, :next_run_at, :cron, :cron_description

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
    @schedule_key, @schedule_display, @cron_description = format_schedule(run_at: run_at_val, timezone: @timezone)

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

  def notification_channels
    if @beep&.notification_channels.present?
      @beep.notification_channels
    elsif @params[:notification_channels].present?
      Array(@params[:notification_channels])
    else
      @beep&.assign_default_notification_channels
      @beep&.notification_channels || []
    end
  end

  private
    def beep_attributes
      raw = if @params.respond_to?(:permit)
        @params.permit(:title, :body, :run_at, :cron, :kind, :intent, notification_channels: [], metadata: {})
      else
        @params
      end
      raw.to_h.symbolize_keys.slice(:title, :body, :run_at, :cron, :kind, :intent, :notification_channels, :metadata)
    end

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

    def format_schedule(run_at:, timezone:)
      tz = Time.find_zone(timezone) || Time.find_zone("UTC")
      local_now = Time.current.in_time_zone(tz)

      if @kind == "recurring"
        key = "Cron"
        cron_str = @cron.to_s.strip
        cron_desc = describe_cron(cron_str)
        if @next_run_at.present?
          local_next = @next_run_at.in_time_zone(tz)
          next_str = if local_next.to_date == local_now.to_date
            "Today at #{local_next.strftime('%H:%M')} (#{local_next.strftime('%Y-%m-%d %H:%M:%S')} #{timezone})"
          elsif local_next.to_date == (local_now.to_date + 1.day)
            "Tomorrow at #{local_next.strftime('%H:%M')} (#{local_next.strftime('%Y-%m-%d %H:%M:%S')} #{timezone})"
          else
            "#{local_next.strftime('%Y-%m-%d %H:%M:%S')} #{timezone}"
          end
          display = if cron_desc.present?
            "#{cron_str} (#{cron_desc} · Next: #{next_str})"
          else
            "#{cron_str} (Next: #{next_str})"
          end
        else
          display = cron_desc.present? ? "#{cron_str} (#{cron_desc})" : cron_str
        end
        [ key, display, cron_desc ]
      elsif @params[:in].present?
        key = "Delay"
        in_val = @params[:in].to_s.strip
        if run_at.present?
          local_target = run_at.in_time_zone(tz)
          time_part = if local_target.to_date == local_now.to_date
            "Today at #{local_target.strftime('%H:%M:%S')} · #{local_target.strftime('%Y-%m-%d %H:%M:%S')} #{timezone}"
          elsif local_target.to_date == (local_now.to_date + 1.day)
            "Tomorrow at #{local_target.strftime('%H:%M:%S')} · #{local_target.strftime('%Y-%m-%d %H:%M:%S')} #{timezone}"
          else
            "#{local_target.strftime('%Y-%m-%d %H:%M:%S')} #{timezone}"
          end
          [ key, "#{in_val} (in #{in_val} · #{time_part})", nil ]
        else
          [ key, in_val, nil ]
        end
      elsif @params[:at].present? || @params[:run_at].present? || run_at.present?
        key = "Run At"
        at_val = (@params[:at].presence || @params[:run_at].presence || run_at&.in_time_zone(tz)&.strftime("%Y-%m-%d %H:%M")).to_s.strip
        if run_at.present?
          local_target = run_at.in_time_zone(tz)
          display = if local_target.to_date == local_now.to_date
            "#{at_val} (Today at #{local_target.strftime('%H:%M')} · #{local_target.strftime('%Y-%m-%d %H:%M:%S')} #{timezone})"
          elsif local_target.to_date == (local_now.to_date + 1.day)
            "#{at_val} (Tomorrow at #{local_target.strftime('%H:%M')} · #{local_target.strftime('%Y-%m-%d %H:%M:%S')} #{timezone})"
          else
            diff_days = (local_target.to_date - local_now.to_date).to_i
            "#{at_val} (#{local_target.strftime('%Y-%m-%d %H:%M:%S')} #{timezone} · in #{diff_days} days)"
          end
          [ key, display, nil ]
        else
          [ key, at_val, nil ]
        end
      else
        [ "Schedule", "instant (fires immediately)", nil ]
      end
    end

    def describe_cron(expr)
      return "" if expr.blank?

      clean = expr.strip
      lower = clean.downcase
      return clean.capitalize if lower.start_with?("every ")

      fields = clean.split(/\s+/)
      return "" unless fields.size == 5

      m, h, dom, mon, dow = fields

      if m.start_with?("*/") && h == "*" && dom == "*" && mon == "*" && dow == "*"
        return "Every #{m[2..]} minutes"
      end
      if m == "0" && h == "*" && dom == "*" && mon == "*" && dow == "*"
        return "Every hour"
      end

      if m =~ /\A\d+\z/ && h =~ /\A\d+\z/ && dom == "*" && mon == "*"
        time_str = sprintf("%02d:%02d", h.to_i, m.to_i)
        case dow
        when "*"
          return "Every day at midnight (00:00)" if h.to_i == 0 && m.to_i == 0
          return "Every day at noon (12:00)" if h.to_i == 12 && m.to_i == 0
          "Every day at #{time_str}"
        when "1-5" then "Every weekday at #{time_str}"
        when "0,6", "6,0", "7,6", "6,7" then "Every weekend at #{time_str}"
        when "1" then "Every Monday at #{time_str}"
        when "2" then "Every Tuesday at #{time_str}"
        when "3" then "Every Wednesday at #{time_str}"
        when "4" then "Every Thursday at #{time_str}"
        when "5" then "Every Friday at #{time_str}"
        when "6" then "Every Saturday at #{time_str}"
        when "0", "7" then "Every Sunday at #{time_str}"
        end
      end
    end
end
