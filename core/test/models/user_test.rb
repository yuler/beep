require "test_helper"

class UserTest < ActiveSupport::TestCase
  test "defaults notification channels to email" do
    user = User.new(name: "Newcomer")

    assert_equal %w[ email ], user.notification_channels
  end

  test "allows an empty notification channel list" do
    user = users(:john)
    user.notification_channels = []

    assert user.valid?
  end

  test "rejects unknown notification channels" do
    user = users(:john)
    user.notification_channels = %w[ sms ]

    assert_not user.valid?
    assert user.errors[:notification_channels].any?
  end

  test "accepts a blank timezone" do
    user = users(:john)
    user.timezone = nil
    user.timezone_source = nil

    assert user.valid?
  end

  test "rejects an unknown timezone" do
    user = users(:john)
    user.timezone = "Not/A_Zone"
    user.timezone_source = "detected"

    assert_not user.valid?
    assert user.errors[:timezone].any?
  end

  test "assign_timezone detected writes only when empty" do
    user = users(:john)
    user.assign_timezone(name: "Asia/Shanghai", source: "detected")
    user.save!

    assert_equal "Asia/Shanghai", user.timezone
    assert_equal "detected", user.timezone_source

    user.assign_timezone(name: "America/New_York", source: "detected")
    user.save!

    assert_equal "Asia/Shanghai", user.timezone
    assert_equal "detected", user.timezone_source
  end

  test "assign_timezone detected does not overwrite manual" do
    user = users(:john)
    user.assign_timezone(name: "Europe/London", source: "manual")
    user.save!

    user.assign_timezone(name: "Asia/Shanghai", source: "detected")
    user.save!

    assert_equal "Europe/London", user.timezone
    assert_equal "manual", user.timezone_source
  end

  test "assign_timezone manual overwrites a detected value" do
    user = users(:john)
    user.assign_timezone(name: "Asia/Shanghai", source: "detected")
    user.save!

    user.assign_timezone(name: "America/New_York", source: "manual")
    user.save!

    assert_equal "America/New_York", user.timezone
    assert_equal "manual", user.timezone_source
  end
end
