require "test_helper"

class Channel::WebPushTest < ActiveSupport::TestCase
  setup do
    stub_web_push_dns_resolution
    @account = accounts(:john_account)
    @user = users(:john)
    @channel = @account.channels.build(
      user: @user,
      kind: :web_push,
      name: "chrome",
      endpoint: "https://fcm.googleapis.com/fcm/send/abc123",
      p256dh_key: "test_key",
      auth_key: "test_auth"
    )
  end

  test "validates web push channel config" do
    assert @channel.valid?

    @channel.endpoint = "http://fcm.googleapis.com/fcm/send/abc123"
    assert_not @channel.valid?
    assert_includes @channel.errors[:endpoint], "must use HTTPS"
  end

  test "deliver_test! sends push notification" do
    @channel.save!
    sent = false
    stub_web_push_payload_send(->(**kwargs) {
      sent = true
      assert_equal @channel.endpoint, kwargs[:endpoint]
    }) do
      @channel.deliver_test!
    end

    assert sent
  end

  private
    def stub_web_push_payload_send(callable)
      singleton = WebPush.singleton_class
      singleton.alias_method :__orig_payload_send, :payload_send
      singleton.define_method(:payload_send) { |**kwargs| callable.call(**kwargs) }
      yield
    ensure
      singleton.alias_method :payload_send, :__orig_payload_send
      singleton.remove_method :__orig_payload_send
    end
end
