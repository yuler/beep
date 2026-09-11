require "test_helper"

class Channel::CliTest < ActiveSupport::TestCase
  setup do
    @account = accounts(:john_account)
    @user = users(:john)
    @channel = @account.channels.create!(
      user: @user,
      kind: :cli,
      name: "laptop"
    )
    @beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Drink water",
      run_at: 1.hour.from_now
    )
    @run = @beep.runs.create!(scheduled_for: Time.current, status: :pending)
  end

  test "deliver_beep creates channel delivery" do
    result = @channel.deliver_beep(@beep, run: @run)

    assert_equal @channel.id, result["channel_id"]
    assert_equal "laptop", result["channel_name"]
    assert_equal "queued", result["status"]
    assert_predicate result["delivery_id"], :present?

    delivery = @channel.deliveries.find(result["delivery_id"])
    assert_equal "pending", delivery.status
    assert_equal "Drink water", delivery.payload["title"]
  end
end
