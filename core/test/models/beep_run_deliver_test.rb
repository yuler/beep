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
    ActionMailer::Base.deliveries.clear
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
    assert_equal "sent", @run.result.dig("email", "status")
    assert_equal 1, ActionMailer::Base.deliveries.size
    assert_equal [ "john@example.com" ], ActionMailer::Base.deliveries.last.to
  end

  test "deliver sends web push to the recipient user's subscriptions" do
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
    assert_equal "Call mom", payload["title"]
    assert_nil payload.dig("options", "body")
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
    ActionMailer::Base.deliveries.clear

    stub_web_push_payload_send(->(**_kwargs) { sent += 1 }) do
      @run.deliver_now
    end

    assert_equal 0, sent
    assert_equal 0, ActionMailer::Base.deliveries.size
  end

  test "deliver retries email without sending web push again" do
    subscribe("https://fcm.googleapis.com/fcm/send/one")
    sent = 0
    stub_web_push_payload_send(->(**_kwargs) { sent += 1 }) do
      fail_email_delivery do
        assert_raises BeepRun::EmailDeliveryError do
          @run.deliver_now
        end
      end
    end

    assert_equal 1, sent
    assert @run.reload.running?
    assert_equal "error", @run.result.dig("email", "status")

    sent = 0
    stub_web_push_payload_send(->(**_kwargs) { sent += 1 }) do
      @run.deliver_now
    end

    assert_equal 0, sent
    assert @run.reload.succeeded?
    assert_equal "sent", @run.result.dig("email", "status")
    assert_equal 1, ActionMailer::Base.deliveries.size
  end

  test "deliver skips email when the beep lists only web_push" do
    @beep.update!(notification_channels: %w[ web_push ])

    @run.deliver_now

    assert @run.reload.succeeded?
    assert_nil @run.result["email"]
    assert_equal "no_subscriptions", @run.result.dig("web_push", "reason")
    assert_equal 0, ActionMailer::Base.deliveries.size
  end

  test "deliver skips web push when the beep lists only email" do
    subscribe("https://fcm.googleapis.com/fcm/send/one")
    @beep.update!(notification_channels: %w[ email ])
    sent = 0

    stub_web_push_payload_send(->(**_kwargs) { sent += 1 }) do
      @run.deliver_now
    end

    assert_equal 0, sent
    assert @run.reload.succeeded?
    assert_nil @run.result["web_push"]
    assert_equal "sent", @run.result.dig("email", "status")
  end

  test "deliver succeeds with an empty result when the beep has no channels" do
    @beep.update!(notification_channels: [])

    @run.deliver_now

    assert @run.reload.succeeded?
    assert_equal({}, @run.result)
    assert_equal 0, ActionMailer::Base.deliveries.size
  end

  test "deliver emails a team owner when email is in their channels" do
    team = Account.create_with_owner(
      account: { name: "John Team", personal: false, slug: "john_tdeliv" },
      owner: { name: "John", identity: identities(:john) }
    )
    team.users.find_by!(role: :owner).update!(notification_channels: %w[ email web_push ])
    member = team.users.create!(
      name: "Yuler",
      identity: identities(:yuler),
      role: :member,
      verified_at: Time.current
    )
    beep = due_once_beep(account: team)
    Beep.poll_due_now
    run = beep.runs.sole
    subscribe("https://fcm.googleapis.com/fcm/send/owner", user: team.users.find_by!(role: :owner))
    subscribe("https://fcm.googleapis.com/fcm/send/member", user: member)
    sent_endpoints = []

    stub_web_push_payload_send(->(**kwargs) {
      sent_endpoints << kwargs[:endpoint]
    }) do
      run.deliver_now
    end

    assert run.reload.succeeded?
    assert_equal [ "https://fcm.googleapis.com/fcm/send/owner" ], sent_endpoints
    assert_equal "sent", run.result.dig("email", "status")
    assert_equal [ "john@example.com" ], ActionMailer::Base.deliveries.last.to
  end

  test "deliver is a no-op for an expired run" do
    expired_run = BeepRun.create!(beep: @beep, scheduled_for: 1.hour.ago, status: :expired)

    expired_run.deliver_now

    assert expired_run.reload.expired?
    assert_nil expired_run.result
  end

  private
    def due_once_beep(account: @account)
      beep = Beep.create!(
        account: account,
        kind: :once,
        title: "Call mom",
        run_at: 1.hour.from_now.change(usec: 0)
      )
      beep.update_columns(next_run_at: 1.minute.ago.change(usec: 0))
      beep
    end

    def subscribe(endpoint, user: users(:john))
      Push::Subscription.new(
        user: user,
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

    def fail_email_delivery
      previous = ActionMailer::Base.delivery_method
      ActionMailer::Base.add_delivery_method :failing, FailingDelivery
      ActionMailer::Base.delivery_method = :failing
      yield
    ensure
      ActionMailer::Base.delivery_method = previous
    end
end

class FailingDelivery
  def initialize(*)
  end

  def deliver!(*)
    raise Net::ReadTimeout
  end
end
