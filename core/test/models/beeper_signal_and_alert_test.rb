require "test_helper"

class BeeperSignalAndAlertTest < ActiveSupport::TestCase
  include ActiveJob::TestHelper

  setup do
    stub_web_push_dns_resolution
    @account = accounts(:john_account)
    @beeper_app = BeeperApp.create!(
      slug: "echo",
      version: "1.0.0",
      manifest: {
        "manifest_version" => 1,
        "slug" => "echo",
        "name" => "Echo Monitor",
        "version" => "1.0.0",
        "author" => "Beep",
        "schedule" => {
          "default_cron" => "*/5 * * * *"
        },
        "alerting" => {
          "failure_threshold" => 2
        }
      }
    )

    @beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      cron: "*/5 * * * *",
      timezone: "UTC",
      title: "Echo Signal Monitor",
      config: { "status" => "ok" },
      notification_channels: %w[ email ]
    )

    ActionMailer::Base.deliveries.clear
  end

  test "healthy signal stays silent and updates signal_status to ok" do
    assert_no_difference -> { @account.beeps.count } do
      run = @beeper.runs.create!(scheduled_for: Time.current, status: :pending)
      run.execute_now

      run.reload
      @beeper.reload

      assert_equal "succeeded", run.status
      assert_equal "ok", run.signal_status
      assert_equal "ok", @beeper.alert_state
      assert_equal 0, @beeper.consecutive_failures
      assert_empty ActionMailer::Base.deliveries
    end
  end

  test "single failure below threshold does not notify, sets alert_state to pending, and increments consecutive_failures" do
    @beeper.update!(config: { "status" => "alerting", "title" => "Target Down", "message" => "HTTP 500" })

    assert_no_difference -> { @account.beeps.count } do
      run = @beeper.runs.create!(scheduled_for: Time.current, status: :pending)
      run.execute_now

      run.reload
      @beeper.reload

      assert_equal "succeeded", run.status
      assert_equal "alerting", run.signal_status
      assert_equal "pending", @beeper.alert_state
      assert_equal 1, @beeper.consecutive_failures
      assert_empty ActionMailer::Base.deliveries
    end
  end

  test "reaching failure threshold transitions from pending to alerting and triggers alert notification" do
    @beeper.update!(
      alert_state: :pending,
      consecutive_failures: 1,
      config: { "status" => "alerting", "title" => "Target Down", "message" => "HTTP 500" }
    )

    assert_difference -> { @account.beeps.count }, 1 do
      run = @beeper.runs.create!(scheduled_for: Time.current, status: :pending)
      run.execute_now

      run.reload
      @beeper.reload

      assert_equal "succeeded", run.status
      assert_equal "alerting", run.signal_status
      assert_equal "alerting", @beeper.alert_state
      assert_equal 2, @beeper.consecutive_failures

      created_beep = @account.beeps.order(:created_at).last
      assert_equal "once", created_beep.kind
      assert_equal "Target Down", created_beep.title
      assert_equal "HTTP 500", created_beep.body
      assert_equal @beeper.id, created_beep.beeper_id
      assert_equal [ "email" ], created_beep.notification_channels

      # Deliver the created once beep run
      beep_run = created_beep.runs.sole
      beep_run.deliver_now

      assert_equal 1, ActionMailer::Base.deliveries.size
      assert_equal "Target Down", ActionMailer::Base.deliveries.last.subject
    end
  end

  test "subsequent failures while already alerting do not spam notifications" do
    @beeper.update!(
      alert_state: :alerting,
      consecutive_failures: 2,
      config: { "status" => "alerting", "title" => "Target Down", "message" => "HTTP 500" }
    )

    assert_no_difference -> { @account.beeps.count } do
      run = @beeper.runs.create!(scheduled_for: Time.current, status: :pending)
      run.execute_now

      run.reload
      @beeper.reload

      assert_equal "succeeded", run.status
      assert_equal "alerting", @beeper.alert_state
      assert_equal 3, @beeper.consecutive_failures
      assert_empty ActionMailer::Base.deliveries
    end
  end

  test "recovery from alerting with default recovery_threshold of 1 creates recovery Beep and resets alert_state" do
    @beeper.update!(
      alert_state: :alerting,
      consecutive_failures: 3,
      config: { "status" => "ok", "title" => "Target Recovered", "message" => "Service back online" }
    )

    assert_difference -> { @account.beeps.count }, 1 do
      run = @beeper.runs.create!(scheduled_for: Time.current, status: :pending)
      run.execute_now

      run.reload
      @beeper.reload

      assert_equal "succeeded", run.status
      assert_equal "ok", run.signal_status
      assert_equal "ok", @beeper.alert_state
      assert_equal 0, @beeper.consecutive_failures
      assert_equal 0, @beeper.consecutive_recoveries

      created_beep = @account.beeps.order(:created_at).last
      assert_equal "once", created_beep.kind
      assert_equal "Target Recovered", created_beep.title
      assert_equal @beeper.id, created_beep.beeper_id

      # Deliver the created once beep run
      beep_run = created_beep.runs.sole
      beep_run.deliver_now

      assert_equal 1, ActionMailer::Base.deliveries.size
      assert_equal "Target Recovered", ActionMailer::Base.deliveries.last.subject
    end
  end

  test "recovery hysteresis with recovery_threshold > 1 transitions through recovering state before ok" do
    @beeper.update!(
      config: { "status" => "ok", "title" => "Target Recovered", "message" => "Service back online", "alerting" => { "recovery_threshold" => 2 } },
      alert_state: :alerting,
      consecutive_failures: 2
    )

    # First ok run -> transitions to recovering, no notification yet
    assert_no_difference -> { @account.beeps.count } do
      run1 = @beeper.runs.create!(scheduled_for: Time.current, status: :pending)
      run1.execute_now
      @beeper.reload
      assert_equal "recovering", @beeper.alert_state
      assert_equal 1, @beeper.consecutive_recoveries
    end

    # Second ok run -> reaches recovery_threshold (2), triggers recovery notification and resets to ok
    assert_difference -> { @account.beeps.count }, 1 do
      run2 = @beeper.runs.create!(scheduled_for: Time.current, status: :pending)
      run2.execute_now
      @beeper.reload
      assert_equal "ok", @beeper.alert_state
      assert_equal 0, @beeper.consecutive_recoveries
      assert_equal 0, @beeper.consecutive_failures
    end
  end

  test "windowed alert policy triggers alert when failure count within window reaches min_failures" do
    @beeper.update!(
      config: {
        "status" => "alerting",
        "title" => "Flapping Target",
        "message" => "Intermittent failure",
        "alerting" => { "policy" => "windowed", "window_size" => 3, "min_failures" => 2 }
      }
    )

    # 1st failure: 1/3 in window -> pending
    assert_no_difference -> { @account.beeps.count } do
      run1 = @beeper.runs.create!(scheduled_for: 2.minutes.ago, status: :pending)
      run1.execute_now
      assert_equal "pending", @beeper.reload.alert_state
    end

    # 2nd failure: 2/3 in window -> triggers alerting Beep
    assert_difference -> { @account.beeps.count }, 1 do
      run2 = @beeper.runs.create!(scheduled_for: 1.minute.ago, status: :pending)
      run2.execute_now
      assert_equal "alerting", @beeper.reload.alert_state
    end
  end

  test "trigger_run! on a paused beeper preserves paused status after run completes" do
    @beeper.pause!
    assert @beeper.paused?

    run = @beeper.trigger_run!
    assert @beeper.reload.paused?

    run.execute_now
    assert @beeper.reload.paused?
  end

  test "notify_from! truncates long titles and messages to fit Beep constraints" do
    long_title = "A" * 150
    long_message = "B" * 3000
    signal = BeeperApp::Signal.new(status: :alerting, title: long_title, message: long_message)

    assert_difference -> { @account.beeps.count }, 1 do
      @beeper.notify_from!(signal)
    end

    created_beep = @account.beeps.order(:created_at).last
    assert_equal Beep::TITLE_MAX_LENGTH, created_beep.title.length
    assert_equal Beep::BODY_MAX_LENGTH, created_beep.body.length
  end

  test "validates config inputs against beeper_app manifest inputs" do
    manifest_app = BeeperApp.create!(
      slug: "custom-probe",
      version: "1.0.0",
      manifest: {
        "manifest_version" => 1,
        "slug" => "custom-probe",
        "name" => "Custom Probe",
        "version" => "1.0.0",
        "author" => "Beep",
        "schedule" => { "default_cron" => "*/5 * * * *" },
        "inputs" => [
          { "name" => "endpoint", "label" => "Endpoint", "type" => "url", "required" => true },
          { "name" => "retries", "label" => "Retries", "type" => "number", "min" => 1, "max" => 5 },
          { "name" => "mode", "label" => "Mode", "type" => "enum", "options" => %w[ fast deep ] }
        ]
      }
    )

    # Missing required endpoint
    invalid_beeper = Beeper.new(
      account: @account,
      beeper_app: manifest_app,
      title: "Test Probe",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: {}
    )
    assert_not invalid_beeper.valid?
    assert_includes invalid_beeper.errors[:config], "Endpoint is required"

    # Number out of range
    invalid_beeper.config = { "endpoint" => "https://example.com", "retries" => 10 }
    assert_not invalid_beeper.valid?
    assert_includes invalid_beeper.errors[:config], "Retries must be at most 5"

    # Invalid enum
    invalid_beeper.config = { "endpoint" => "https://example.com", "retries" => 3, "mode" => "invalid" }
    assert_not invalid_beeper.valid?
    assert_includes invalid_beeper.errors[:config], "Mode must be one of: fast, deep"

    # Valid config
    valid_beeper = Beeper.new(
      account: @account,
      beeper_app: manifest_app,
      title: "Test Probe",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "endpoint" => "https://example.com/api", "retries" => 3, "mode" => "fast" }
    )
    assert valid_beeper.valid?
  end

  test "beeper validates cron against beeper_app min_interval_seconds" do
    app_with_min_interval = BeeperApp.create!(
      account: @account,
      slug: "custom-hourly",
      version: "1.0.0",
      manifest: {
        "manifest_version" => 1,
        "slug" => "custom-hourly",
        "name" => "Custom Hourly",
        "version" => "1.0.0",
        "author" => "Me",
        "schedule" => {
          "default_cron" => "0 * * * *",
          "min_interval_seconds" => 300 # 5 minutes
        }
      }
    )

    # 1 minute cron should fail
    beeper = Beeper.new(
      account: @account,
      beeper_app: app_with_min_interval,
      title: "Too Fast",
      cron: "* * * * *",
      timezone: "UTC",
      config: {}
    )
    assert_not beeper.valid?
    assert_includes beeper.errors[:cron], "interval cannot be shorter than 300 seconds"

    # 5 minutes cron should pass
    beeper.cron = "*/5 * * * *"
    assert beeper.valid?
  end

  test "sanitize_signal_result truncates oversized messages and trims large metrics" do
    run = @beeper.runs.new(scheduled_for: Time.current, status: :pending)

    huge_hash = {
      "status" => "ok",
      "title" => "T" * 300,
      "message" => "M" * 10_000,
      "metrics" => (1..50).to_h { |i| [ "metric_#{i}", "value_#{i}_" + ("X" * 200) ] }
    }

    sanitized = run.send(:sanitize_signal_result, huge_hash)
    assert sanitized["truncated"]
    assert sanitized["message"].length <= 503 # 500 + '...'
    assert sanitized["title"].length <= 203
    assert sanitized["metrics"].size <= 20
    assert sanitized.to_json.bytesize <= BeeperRun::SIGNAL_RESULT_MAX_BYTES
  end
end
