require "test_helper"

class BeepRunDeliverTest < ActiveSupport::TestCase
  setup do
    stub_web_push_dns_resolution
    @account = accounts(:john_account)
    @beep = due_once_beep
    Beep.poll_due_now
    @run = @beep.runs.sole
    @previous_public_key = Rails.application.config.x.vapid.public_key
    @previous_private_key = Rails.application.config.x.vapid.private_key
    Rails.application.config.x.vapid.public_key = "test-vapid-public"
    Rails.application.config.x.vapid.private_key = "test-vapid-private"
  end

  teardown do
    Rails.application.config.x.vapid.public_key = @previous_public_key
    Rails.application.config.x.vapid.private_key = @previous_private_key
  end

  test "deliver with no subscriptions succeeds and completes the beep" do
    @run.deliver_now

    @run.reload
    @beep.reload
    assert @run.succeeded?
    assert_equal "no_subscriptions", @run.result.dig("web_push", "reason")
    assert @beep.completed?
    assert_nil @beep.next_run_at
    assert_equal @run.scheduled_for.to_i, @beep.last_run_at.to_i
  end

  test "deliver sends web push to every subscription on the account" do
    first = subscribe("https://fcm.googleapis.com/fcm/send/one")
    second = subscribe("https://fcm.googleapis.com/fcm/send/two")
    sent_endpoints = []
    payload = nil

    stub_web_push_payload_send(->(**kwargs) {
      sent_endpoints << kwargs[:endpoint]
      payload = JSON.parse(kwargs[:message])
    }) do
      @run.deliver_now
    end

    assert_equal [ first.endpoint, second.endpoint ].sort, sent_endpoints.sort
    assert_equal "Beep", payload["title"]
    assert_equal "Call mom", payload.dig("options", "body")
    assert_equal "#{Rails.application.config.x.web_origin}/#{@account.slug}/beeps/#{@beep.id}", payload.dig("options", "data", "url")
    assert @run.reload.succeeded?
    statuses = @run.result.dig("web_push", "deliveries").map { |row| row["status"] }
    assert_equal [ "sent", "sent" ], statuses
    assert @beep.reload.completed?
  end

  test "deliver deletes expired subscriptions and still succeeds" do
    live = subscribe("https://fcm.googleapis.com/fcm/send/live")
    expired = subscribe("https://fcm.googleapis.com/fcm/send/expired")
    push_response = Object.new
    def push_response.body = "Gone"
    gone = WebPush::ExpiredSubscription.new(push_response, "fcm.googleapis.com")

    stub_web_push_payload_send(->(**kwargs) {
      raise gone if kwargs[:endpoint] == expired.endpoint
    }) do
      @run.deliver_now
    end

    assert @run.reload.succeeded?
    assert Push::Subscription.exists?(live.id)
    assert_not Push::Subscription.exists?(expired.id)
    deliveries = @run.result.dig("web_push", "deliveries").index_by { |row| row["subscription_id"] }
    assert_equal "sent", deliveries[live.id]["status"]
    assert_equal "expired", deliveries[expired.id]["status"]
    assert @beep.reload.completed?
  end

  test "deliver records per-device errors without failing the run" do
    subscribe("https://fcm.googleapis.com/fcm/send/ok")
    bad = subscribe("https://fcm.googleapis.com/fcm/send/bad")

    stub_web_push_payload_send(->(**kwargs) {
      raise Timeout::Error if kwargs[:endpoint] == bad.endpoint
    }) do
      @run.deliver_now
    end

    assert @run.reload.succeeded?
    deliveries = @run.result.dig("web_push", "deliveries").index_by { |row| row["subscription_id"] }
    assert_equal "sent", deliveries.except(bad.id).values.first["status"]
    assert_equal "error", deliveries[bad.id]["status"]
    assert_equal "Timeout::Error", deliveries[bad.id]["error"]
    assert @beep.reload.completed?
  end

  test "deliver is a no-op when the run already succeeded" do
    @run.deliver_now
    sent = 0

    stub_web_push_payload_send(->(**_kwargs) { sent += 1 }) do
      @run.deliver_now
    end

    assert_equal 0, sent
  end

  test "deliver is a no-op when the run is already running" do
    @run.update_columns(status: "running")
    sent = 0

    stub_web_push_payload_send(->(**_kwargs) { sent += 1 }) do
      @run.deliver_now
    end

    assert_equal 0, sent
    assert @run.reload.running?
  end

  test "deliver is a no-op for an expired run" do
    expired_run = BeepRun.create!(beep: @beep, scheduled_for: 1.hour.ago, status: :expired)

    expired_run.deliver_now

    assert expired_run.reload.expired?
    assert_nil expired_run.result
  end

  private
    def due_once_beep
      beep = Beep.create!(
        account: @account,
        kind: :once,
        message: "Call mom",
        run_at: 1.hour.from_now.change(usec: 0)
      )
      beep.update_columns(next_run_at: 1.minute.ago.change(usec: 0))
      beep
    end

    def subscribe(endpoint)
      Push::Subscription.new(
        user: users(:john),
        endpoint: endpoint,
        p256dh_key: "key",
        auth_key: "auth"
      ).tap(&:save!)
    end

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
