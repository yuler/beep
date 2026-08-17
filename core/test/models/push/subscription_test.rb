require "test_helper"

class Push::SubscriptionTest < ActiveSupport::TestCase
  setup do
    stub_web_push_dns_resolution
  end

  test "valid subscription with permitted endpoint" do
    subscription = Push::Subscription.new(
      user: users(:john),
      endpoint: "https://fcm.googleapis.com/fcm/send/abc123",
      p256dh_key: "test_key",
      auth_key: "test_auth"
    )

    assert subscription.valid?
  end

  test "rejects endpoint with non-https scheme" do
    subscription = Push::Subscription.new(
      user: users(:john),
      endpoint: "http://fcm.googleapis.com/fcm/send/abc123",
      p256dh_key: "test_key",
      auth_key: "test_auth"
    )

    assert_not subscription.valid?
    assert_includes subscription.errors[:endpoint], "must use HTTPS"
  end

  test "rejects endpoint with non-permitted host" do
    subscription = Push::Subscription.new(
      user: users(:john),
      endpoint: "https://attacker.example.com/webhook",
      p256dh_key: "test_key",
      auth_key: "test_auth"
    )

    assert_not subscription.valid?
    assert_includes subscription.errors[:endpoint], "is not a permitted push service"
  end

  test "resolved_endpoint_ip is nil for a private IP" do
    stub_dns_resolution("192.168.1.1")

    subscription = Push::Subscription.new(
      user: users(:john),
      endpoint: "https://fcm.googleapis.com/fcm/send/abc123",
      p256dh_key: "test_key",
      auth_key: "test_auth"
    )

    assert_nil subscription.resolved_endpoint_ip
  end

  test "resolved_endpoint_ip is nil for a loopback IP" do
    stub_dns_resolution("127.0.0.1")

    subscription = Push::Subscription.new(
      user: users(:john),
      endpoint: "https://fcm.googleapis.com/fcm/send/abc123",
      p256dh_key: "test_key",
      auth_key: "test_auth"
    )

    assert_nil subscription.resolved_endpoint_ip
  end

  test "resolved_endpoint_ip is nil for a link-local IP (AWS IMDS)" do
    stub_dns_resolution("169.254.169.254")

    subscription = Push::Subscription.new(
      user: users(:john),
      endpoint: "https://fcm.googleapis.com/fcm/send/abc123",
      p256dh_key: "test_key",
      auth_key: "test_auth"
    )

    assert_nil subscription.resolved_endpoint_ip
  end

  test "resolved_endpoint_ip returns pinned public IP" do
    subscription = Push::Subscription.new(
      user: users(:john),
      endpoint: "https://fcm.googleapis.com/fcm/send/abc123",
      p256dh_key: "test_key",
      auth_key: "test_auth"
    )

    assert_equal DnsTestHelper::WEB_PUSH_PUBLIC_TEST_IP, subscription.resolved_endpoint_ip
  end

  test "accepts all permitted push service domains" do
    permitted_endpoints = [
      "https://fcm.googleapis.com/fcm/send/token123",
      "https://jmt17.google.com/fcm/send/token123",
      "https://updates.push.services.mozilla.com/wpush/v2/token123",
      "https://web.push.apple.com/QaBC123",
      "https://wns2-db5p.notify.windows.com/w/?token=abc123"
    ]

    permitted_endpoints.each do |endpoint|
      subscription = Push::Subscription.new(
        user: users(:john),
        endpoint: endpoint,
        p256dh_key: "test_key",
        auth_key: "test_auth"
      )

      assert subscription.valid?, "Expected #{endpoint} to be valid, got errors: #{subscription.errors.full_messages}"
    end
  end

  test "upsert_for! creates then updates the same endpoint" do
    attributes = {
      endpoint: "https://fcm.googleapis.com/fcm/send/abc123",
      p256dh_key: "key",
      auth_key: "auth",
      user_agent: "Chrome"
    }

    created = Push::Subscription.upsert_for!(users(:john), attributes)
    updated = Push::Subscription.upsert_for!(users(:john), attributes.merge(p256dh_key: "new-key"))

    assert_equal created.id, updated.id
    assert_equal 1, users(:john).push_subscriptions.where(endpoint: attributes[:endpoint]).count
    assert_equal "new-key", updated.p256dh_key
  end
end
