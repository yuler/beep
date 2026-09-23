module Beep::ScheduleParser
  class Error < StandardError; end

  class << self
    def resolve_run_at(in_val: nil, at_val: nil, run_at_val: nil, timezone: nil, now: Time.current)
      provided = []
      provided << "run_at" if run_at_val.present?
      provided << "in" if in_val.present?
      provided << "at" if at_val.present?

      if provided.size > 1
        raise Error, "Cannot specify multiple schedule options: #{provided.join(', ')}"
      end

      if in_val.present?
        parse_in(in_val, now: now)
      elsif at_val.present?
        parse_at(at_val, timezone: timezone, now: now)
      elsif run_at_val.present?
        run_at_val
      else
        nil
      end
    end

    def parse_in(val, now: Time.current)
      str = val.to_s.strip
      raise Error, "Duration cannot be blank" if str.blank?

      duration = Fugit.parse_duration(str)
      if duration.nil? || duration.to_sec <= 0
        raise Error, "Invalid in duration #{val.inspect} (e.g. 15m, 2h, 1d)"
      end

      now + duration.to_sec.seconds
    end

    def parse_at(val, timezone: nil, now: Time.current)
      str = val.to_s.strip
      raise Error, "Time cannot be blank" if str.blank?

      tz = Time.find_zone(timezone) || Time.find_zone("UTC")
      local_now = now.in_time_zone(tz)

      if str =~ /\A(\d{1,2}):(\d{2})(?::(\d{2}))?\z/
        hour = Regexp.last_match(1).to_i
        minute = Regexp.last_match(2).to_i
        second = (Regexp.last_match(3) || 0).to_i

        unless (0..23).cover?(hour) && (0..59).cover?(minute) && (0..59).cover?(second)
          raise Error, "Invalid at time #{val.inspect}: hour must be 0-23, minute and second 0-59"
        end

        target = tz.local(local_now.year, local_now.month, local_now.day, hour, minute, second)
        target += 1.day if target <= local_now
        return target
      end

      parsed = begin
        tz.parse(str)
      rescue ArgumentError
        nil
      end

      # Time.zone.parse is lenient and coerces arbitrary numbers (e.g. "12345") into ancient dates (e.g. year 0012).
      # Rejecting dates prior to year 2000 prevents false-positive parses for non-date inputs while supporting all modern schedule dates.
      if parsed && parsed.year >= 2000
        if parsed <= local_now
          raise Error, "Invalid at time #{val.inspect}: cannot be in the past"
        end

        return parsed
      end

      raise Error, "Invalid at time #{val.inspect} (e.g. 15:30, 2026-10-01 10:00)"
    end
  end
end
