require "test_helper"

class Beep::ScheduleParserTest < ActiveSupport::TestCase
  test "parse_in parses standard duration units" do
    now = Time.utc(2026, 9, 23, 10, 0, 0)

    assert_equal Time.utc(2026, 9, 23, 10, 15, 0), Beep::ScheduleParser.parse_in("15m", now: now)
    assert_equal Time.utc(2026, 9, 23, 12, 0, 0), Beep::ScheduleParser.parse_in("2h", now: now)
    assert_equal Time.utc(2026, 9, 24, 10, 0, 0), Beep::ScheduleParser.parse_in("1d", now: now)
    assert_equal Time.utc(2026, 9, 23, 10, 0, 30), Beep::ScheduleParser.parse_in("30s", now: now)
    assert_equal Time.utc(2026, 9, 23, 11, 30, 0), Beep::ScheduleParser.parse_in("1h30m", now: now)
  end

  test "parse_in rejects blank or invalid duration" do
    assert_raises(Beep::ScheduleParser::Error) { Beep::ScheduleParser.parse_in("") }
    assert_raises(Beep::ScheduleParser::Error) { Beep::ScheduleParser.parse_in("foo") }
    assert_raises(Beep::ScheduleParser::Error) { Beep::ScheduleParser.parse_in("0s") }
    assert_raises(Beep::ScheduleParser::Error) { Beep::ScheduleParser.parse_in("-5m") }
  end

  test "parse_at parses time-only string and respects timezone" do
    tz = "Asia/Shanghai" # UTC+8
    now = Time.find_zone(tz).local(2026, 9, 23, 10, 0, 0) # 10:00 local

    # Future time today (15:30)
    result = Beep::ScheduleParser.parse_at("15:30", timezone: tz, now: now)
    assert_equal Time.find_zone(tz).local(2026, 9, 23, 15, 30, 0), result

    # Past time today (09:00) rolls over to next day
    result_tomorrow = Beep::ScheduleParser.parse_at("09:00", timezone: tz, now: now)
    assert_equal Time.find_zone(tz).local(2026, 9, 24, 9, 0, 0), result_tomorrow

    # Exact same time rolls over to next day
    result_same = Beep::ScheduleParser.parse_at("10:00", timezone: tz, now: now)
    assert_equal Time.find_zone(tz).local(2026, 9, 24, 10, 0, 0), result_same
  end

  test "parse_at parses full datetime string" do
    tz = "Asia/Shanghai"
    now = Time.find_zone(tz).local(2026, 9, 23, 10, 0, 0)

    result = Beep::ScheduleParser.parse_at("2026-10-01 14:00", timezone: tz, now: now)
    assert_equal Time.find_zone(tz).local(2026, 10, 1, 14, 0, 0), result

    result_iso = Beep::ScheduleParser.parse_at("2026-10-01T14:00:00+08:00", timezone: tz, now: now)
    assert_equal Time.find_zone(tz).local(2026, 10, 1, 14, 0, 0), result_iso
  end

  test "parse_at rejects invalid times or invalid formats" do
    assert_raises(Beep::ScheduleParser::Error) { Beep::ScheduleParser.parse_at("") }
    assert_raises(Beep::ScheduleParser::Error) { Beep::ScheduleParser.parse_at("25:00") }
    assert_raises(Beep::ScheduleParser::Error) { Beep::ScheduleParser.parse_at("12:60") }
    assert_raises(Beep::ScheduleParser::Error) { Beep::ScheduleParser.parse_at("not-a-time") }
    assert_raises(Beep::ScheduleParser::Error) { Beep::ScheduleParser.parse_at("12345") }
  end

  test "parse_at rejects past datetime" do
    tz = "Asia/Shanghai"
    now = Time.find_zone(tz).local(2026, 9, 23, 10, 0, 0)

    err = assert_raises(Beep::ScheduleParser::Error) do
      Beep::ScheduleParser.parse_at("2026-09-20 10:00", timezone: tz, now: now)
    end
    assert_match(/cannot be in the past/, err.message)

    # Exact current datetime is also rejected (must be in the future)
    assert_raises(Beep::ScheduleParser::Error) do
      Beep::ScheduleParser.parse_at("2026-09-23 10:00:00", timezone: tz, now: now)
    end
  end

  test "resolve_run_at handles mutual exclusivity" do
    assert_raises(Beep::ScheduleParser::Error) do
      Beep::ScheduleParser.resolve_run_at(in_val: "15m", at_val: "15:30")
    end

    assert_raises(Beep::ScheduleParser::Error) do
      Beep::ScheduleParser.resolve_run_at(in_val: "15m", run_at_val: "2026-10-01 10:00")
    end

    assert_raises(Beep::ScheduleParser::Error) do
      Beep::ScheduleParser.resolve_run_at(at_val: "15:30", run_at_val: "2026-10-01 10:00")
    end
  end

  test "resolve_run_at returns parsed time for single option" do
    now = Time.utc(2026, 9, 23, 10, 0, 0)

    in_res = Beep::ScheduleParser.resolve_run_at(in_val: "10m", now: now)
    assert_equal Time.utc(2026, 9, 23, 10, 10, 0), in_res

    at_res = Beep::ScheduleParser.resolve_run_at(at_val: "11:00", timezone: "UTC", now: now)
    assert_equal Time.utc(2026, 9, 23, 11, 0, 0), at_res

    run_at_res = Beep::ScheduleParser.resolve_run_at(run_at_val: "2026-10-01T10:00:00Z")
    assert_equal "2026-10-01T10:00:00Z", run_at_res

    assert_nil Beep::ScheduleParser.resolve_run_at
  end
end
