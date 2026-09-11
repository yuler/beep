require "test_helper"

class Channel::EmailTest < ActiveSupport::TestCase
  setup do
    @account = accounts(:john_account)
    @user = users(:john)
    @channel = @account.channels.create!(
      user: @user,
      kind: :email,
      name: "john-email"
    )
    @beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Reminder email",
      run_at: 1.hour.from_now
    )
    @run = @beep.runs.create!(scheduled_for: Time.current, status: :pending)
    ActionMailer::Base.deliveries.clear
  end

  test "deliver_beep sends reminder email" do
    result = @channel.deliver_beep(@beep, run: @run)

    assert_equal @channel.id, result["channel_id"]
    assert_equal "sent", result["status"]
    assert_equal 1, ActionMailer::Base.deliveries.size
    assert_equal [ @user.identity.email ], ActionMailer::Base.deliveries.last.to
  end
end
