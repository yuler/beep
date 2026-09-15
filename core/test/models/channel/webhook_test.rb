require "test_helper"

class Channel::WebhookTest < ActiveSupport::TestCase
  setup do
    @account = accounts(:john_account)
    @user = users(:john)
    @channel = @account.channels.build(
      user: @user,
      kind: :webhook,
      name: "ops-hook"
    )
    @beep = Beep.create!(
      account: @account,
      kind: :once,
      title: "Drink water",
      run_at: 1.hour.from_now
    )
    @run = @beep.runs.create!(scheduled_for: Time.current, status: :pending)
  end

  test "webhook channels are rejected at validation" do
    assert_not @channel.valid?
    assert_includes @channel.errors[:base], "Webhook channels are not supported yet"
  end

  test "deliver_beep raises instead of silently dropping" do
    assert_raises(NotImplementedError) do
      @channel.deliver_beep(@beep, run: @run)
    end
  end

  test "deliver_test! raises instead of reporting success" do
    assert_raises(NotImplementedError) do
      @channel.deliver_test!
    end
  end
end
