require "test_helper"

class PluginOfficialCheckersTest < ActiveSupport::TestCase
  setup do
    stub_web_push_dns_resolution
    @account = accounts(:john_account)
  end

  test "SiteUptime checker reports ok when status matches expected" do
    fake_response = Net::HTTPSuccess.new("1.1", "200", "OK")

    checker = Plugin::Checkers::SiteUptime.new(config: {
      "target_url" => "https://example.com/health",
      "expected_status" => 200
    })

    checker.define_singleton_method(:fetch_with_redirects) do |*, **|
      [ fake_response, URI("https://example.com/health") ]
    end

    result = checker.call
    assert result.ok?
    assert_equal 200, result.metrics["status"]
    assert_match(/operational/, result.title)
  end

  test "SiteUptime checker reports alerting when status does not match" do
    fake_response = Net::HTTPNotFound.new("1.1", "404", "Not Found")

    checker = Plugin::Checkers::SiteUptime.new(config: {
      "target_url" => "https://example.com/health",
      "expected_status" => 200
    })

    checker.define_singleton_method(:fetch_with_redirects) do |*, **|
      [ fake_response, URI("https://example.com/health") ]
    end

    result = checker.call
    assert result.alerting?
    assert_equal 404, result.metrics["status"]
    assert_match(/HTTP 404/, result.title)
  end

  test "SiteUptime blocks private / local IP addresses via SSRF protection" do
    stub_dns_resolution("127.0.0.1")

    result = Plugin::Checkers::SiteUptime.call(config: {
      "target_url" => "http://127.0.0.1:3000"
    })

    assert result.error?
    assert_match(/Blocked target address/, result.title)
  end

  test "SslExpiry checker reports ok when days remaining >= threshold" do
    fake_cert = Struct.new(:not_after).new(30.days.from_now)
    checker = Plugin::Checkers::SslExpiry.new(config: {
      "hostname" => "example.com",
      "alert_days_before" => 14
    })

    stub_dns_resolution("93.184.216.34")
    checker.define_singleton_method(:fetch_peer_certificate) do |*, **|
      fake_cert
    end

    result = checker.call
    assert result.ok?
    assert result.metrics["days_remaining"] >= 29
  end

  test "SslExpiry checker reports alerting when expiring soon" do
    fake_cert = Struct.new(:not_after).new(5.days.from_now)
    checker = Plugin::Checkers::SslExpiry.new(config: {
      "hostname" => "example.com",
      "alert_days_before" => 14
    })

    stub_dns_resolution("93.184.216.34")
    checker.define_singleton_method(:fetch_peer_certificate) do |*, **|
      fake_cert
    end

    result = checker.call
    assert result.alerting?
    assert_match(/expiring soon/, result.title)
  end

  test "Heartbeat checker reports ok when ping is recent" do
    result = Plugin::Checkers::Heartbeat.call(config: {
      "grace_period_minutes" => 15,
      "last_ping_at" => 5.minutes.ago.iso8601
    })

    assert result.ok?
    assert_equal 5, result.metrics["minutes_since_last_ping"]
  end

  test "Heartbeat checker reports alerting when ping is stale" do
    result = Plugin::Checkers::Heartbeat.call(config: {
      "grace_period_minutes" => 15,
      "last_ping_at" => 30.minutes.ago.iso8601
    })

    assert result.alerting?
    assert_match(/missing/, result.title)
    assert_equal 30, result.metrics["minutes_since_last_ping"]
  end

  test "Heartbeat checker reports alerting when no ping was ever received" do
    result = Plugin::Checkers::Heartbeat.call(config: {
      "grace_period_minutes" => 15,
      "last_ping_at" => nil
    })

    assert result.alerting?
    assert_match(/never received/, result.title)
  end
end
