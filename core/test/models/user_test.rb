require "test_helper"

class UserTest < ActiveSupport::TestCase
  test "defaults notification channels to email and web push" do
    user = users(:john)

    assert_equal %w[ email web_push ], user.notification_channels
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
end
