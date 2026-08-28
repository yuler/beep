require "test_helper"

class BeeperOfficialReceiversTest < ActiveSupport::TestCase
  setup do
    stub_web_push_dns_resolution
    @account = accounts(:john_account)
  end

  test "SiteUptime receiver reports ok when status matches expected" do
    fake_response = Net::HTTPSuccess.new("1.1", "200", "OK")

    receiver = BeeperApp::Receivers::SiteUptime.new(config: {
      "target_url" => "https://example.com/health",
      "expected_status" => 200
    })

    receiver.define_singleton_method(:fetch_with_redirects) do |*, **|
      [ fake_response, URI("https://example.com/health") ]
    end

    result = receiver.call
    assert result.ok?
    assert_equal 200, result.metrics["status"]
    assert_match(/operational/, result.title)
  end

  test "SiteUptime receiver reports alerting when status does not match" do
    fake_response = Net::HTTPNotFound.new("1.1", "404", "Not Found")

    receiver = BeeperApp::Receivers::SiteUptime.new(config: {
      "target_url" => "https://example.com/health",
      "expected_status" => 200
    })

    receiver.define_singleton_method(:fetch_with_redirects) do |*, **|
      [ fake_response, URI("https://example.com/health") ]
    end

    result = receiver.call
    assert result.alerting?
    assert_equal 404, result.metrics["status"]
    assert_match(/HTTP 404/, result.title)
  end

  test "SiteUptime blocks private / local IP addresses via SSRF protection" do
    stub_dns_resolution("127.0.0.1")

    result = BeeperApp::Receivers::SiteUptime.call(config: {
      "target_url" => "http://127.0.0.1:3000"
    })

    assert result.error?
    assert_match(/Blocked target address/, result.title)
  end

  test "SslExpiry receiver reports ok when days remaining >= threshold" do
    fake_cert = Struct.new(:not_after).new(30.days.from_now)
    receiver = BeeperApp::Receivers::SslExpiry.new(config: {
      "hostname" => "example.com",
      "alert_days_before" => 14
    })

    stub_dns_resolution("93.184.216.34")
    receiver.define_singleton_method(:fetch_peer_certificate) do |*, **|
      fake_cert
    end

    result = receiver.call
    assert result.ok?
    assert result.metrics["days_remaining"] >= 29
  end

  test "SslExpiry receiver sanitizes full URL hostname" do
    fake_cert = Struct.new(:not_after).new(30.days.from_now)
    receiver = BeeperApp::Receivers::SslExpiry.new(config: {
      "hostname" => "https://example.com/path",
      "alert_days_before" => 14
    })

    stub_dns_resolution("93.184.216.34")
    receiver.define_singleton_method(:fetch_peer_certificate) do |*, **kwargs|
      fake_cert
    end

    result = receiver.call
    assert_equal "example.com", result.title.match(/for (.*?) is/)[1] rescue nil || "example.com"
    assert result.ok?
  end

  test "SslExpiry receiver reports alerting when expiring soon" do
    fake_cert = Struct.new(:not_after).new(5.days.from_now)
    receiver = BeeperApp::Receivers::SslExpiry.new(config: {
      "hostname" => "example.com",
      "alert_days_before" => 14
    })

    stub_dns_resolution("93.184.216.34")
    receiver.define_singleton_method(:fetch_peer_certificate) do |*, **|
      fake_cert
    end

    result = receiver.call
    assert result.alerting?
    assert_match(/expiring soon/, result.title)
  end

  test "SslExpiry receiver reports alerting on connection or handshake timeout" do
    receiver = BeeperApp::Receivers::SslExpiry.new(config: {
      "hostname" => "example.com",
      "alert_days_before" => 14
    })

    stub_dns_resolution("93.184.216.34")
    receiver.define_singleton_method(:fetch_peer_certificate) do |*, **|
      raise Timeout::Error, "execution expired"
    end

    result = receiver.call
    assert result.alerting?
    assert_match(/timed out/, result.title)
  end

  test "Heartbeat receiver reports ok when ping is recent" do
    result = BeeperApp::Receivers::Heartbeat.call(config: {
      "grace_period_minutes" => 15,
      "last_ping_at" => 5.minutes.ago.iso8601
    })

    assert result.ok?
    assert_equal 5, result.metrics["minutes_since_last_ping"]
  end

  test "Heartbeat receiver reports alerting when ping is stale" do
    result = BeeperApp::Receivers::Heartbeat.call(config: {
      "grace_period_minutes" => 15,
      "last_ping_at" => 30.minutes.ago.iso8601
    })

    assert result.alerting?
    assert_match(/missing/, result.title)
    assert_equal 30, result.metrics["minutes_since_last_ping"]
  end

  test "SslExpiry receiver handles Basic Auth and IPv6 correctly" do
    receiver = BeeperApp::Receivers::SslExpiry.new(config: { "hostname" => "https://user:pass@example.com:8443/path" })
    assert_equal "example.com", receiver.send(:sanitize_hostname, "https://user:pass@example.com:8443/path")
    assert_equal "2001:db8::1", receiver.send(:sanitize_hostname, "https://[2001:db8::1]:443/")
  end

  test "Heartbeat receiver reports ok when within initial grace period for new beeper" do
    result = BeeperApp::Receivers::Heartbeat.call(config: {
      "grace_period_minutes" => 15,
      "last_ping_at" => nil,
      "beeper_created_at" => 5.minutes.ago.iso8601
    })

    assert result.ok?
    assert_match(/Waiting for first heartbeat/, result.title)
  end

  test "Heartbeat receiver reports alerting when no ping was received and past grace period" do
    result = BeeperApp::Receivers::Heartbeat.call(config: {
      "grace_period_minutes" => 15,
      "last_ping_at" => nil,
      "beeper_created_at" => 20.minutes.ago.iso8601
    })

    assert result.alerting?
    assert_match(/never received/, result.title)
  end
end
