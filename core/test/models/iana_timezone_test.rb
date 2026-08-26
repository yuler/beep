require "test_helper"

class IanaTimezoneTest < ActiveSupport::TestCase
  test "valid? accepts IANA identifiers" do
    assert IanaTimezone.valid?("Asia/Shanghai")
    assert IanaTimezone.valid?("UTC")
    assert IanaTimezone.valid?("America/New_York")
  end

  test "valid? rejects blank and unknown names" do
    assert_not IanaTimezone.valid?(nil)
    assert_not IanaTimezone.valid?("")
    assert_not IanaTimezone.valid?("Not/A_Zone")
  end

  test "resolve prefers the user timezone" do
    assert_equal "Europe/London", IanaTimezone.resolve("Europe/London", "Asia/Tokyo")
  end

  test "resolve falls back to the request timezone then UTC" do
    assert_equal "Asia/Tokyo", IanaTimezone.resolve(nil, "Asia/Tokyo")
    assert_equal "UTC", IanaTimezone.resolve(nil, nil)
    assert_equal "UTC", IanaTimezone.resolve("Not/A_Zone", "also/bad")
  end
end
