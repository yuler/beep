require "test_helper"

class BeeperSignalAndAlertTest < ActiveSupport::TestCase
  include ActiveJob::TestHelper

  setup do
    stub_web_push_dns_resolution
    @account = accounts(:john_account)
    @beeper = Beeper.create!(
      slug: "echo",
      version: "1.0.0",
      manifest: {
        "manifest_version" => 1,
        "slug" => "echo",
        "name" => "Echo Monitor",
        "version" => "1.0.0",
        "author" => "Beep",
        "schedule" => {
          "default_cron" => "*/5 * * * *",
          "failure_threshold" => 2
        }
      }
    )

    @install = BeeperInstall.create!(
      account: @account,
      beeper: @beeper,
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
      run = @install.runs.create!(scheduled_for: Time.current, status: :pending)
      run.execute_now

      run.reload
      @install.reload

      assert_equal "succeeded", run.status
      assert_equal "ok", run.signal_status
      assert_equal "ok", @install.alert_state
      assert_equal 0, @install.consecutive_failures
      assert_empty ActionMailer::Base.deliveries
    end
  end

  test "single failure below threshold does not notify but increments consecutive_failures" do
    @install.update!(config: { "status" => "alerting", "title" => "Target Down", "message" => "HTTP 500" })

    assert_no_difference -> { @account.beeps.count } do
      run = @install.runs.create!(scheduled_for: Time.current, status: :pending)
      run.execute_now

      run.reload
      @install.reload

      assert_equal "succeeded", run.status
      assert_equal "alerting", run.signal_status
      assert_equal "ok", @install.alert_state
      assert_equal 1, @install.consecutive_failures
      assert_empty ActionMailer::Base.deliveries
    end
  end

  test "reaching failure threshold triggers alert notification creating a once Beep" do
    @install.update!(
      consecutive_failures: 1,
      config: { "status" => "alerting", "title" => "Target Down", "message" => "HTTP 500" }
    )

    assert_difference -> { @account.beeps.count }, 1 do
      run = @install.runs.create!(scheduled_for: Time.current, status: :pending)
      run.execute_now

      run.reload
      @install.reload

      assert_equal "succeeded", run.status
      assert_equal "alerting", run.signal_status
      assert_equal "alerting", @install.alert_state
      assert_equal 2, @install.consecutive_failures

      created_beep = @account.beeps.order(:created_at).last
      assert_equal "once", created_beep.kind
      assert_equal "Target Down", created_beep.title
      assert_equal "HTTP 500", created_beep.body
      assert_equal @install.id, created_beep.beeper_install_id
      assert_equal [ "email" ], created_beep.notification_channels

      # Deliver the created once beep run
      beep_run = created_beep.runs.sole
      beep_run.deliver_now

      assert_equal 1, ActionMailer::Base.deliveries.size
      assert_equal "Target Down", ActionMailer::Base.deliveries.last.subject
    end
  end

  test "subsequent failures while already alerting do not spam notifications" do
    @install.update!(
      alert_state: :alerting,
      consecutive_failures: 2,
      config: { "status" => "alerting", "title" => "Target Down", "message" => "HTTP 500" }
    )

    assert_no_difference -> { @account.beeps.count } do
      run = @install.runs.create!(scheduled_for: Time.current, status: :pending)
      run.execute_now

      run.reload
      @install.reload

      assert_equal "succeeded", run.status
      assert_equal "alerting", @install.alert_state
      assert_equal 3, @install.consecutive_failures
      assert_empty ActionMailer::Base.deliveries
    end
  end

  test "recovery from alerting creates recovery Beep and resets alert_state" do
    @install.update!(
      alert_state: :alerting,
      consecutive_failures: 3,
      config: { "status" => "ok", "title" => "Target Recovered", "message" => "Service back online" }
    )

    assert_difference -> { @account.beeps.count }, 1 do
      run = @install.runs.create!(scheduled_for: Time.current, status: :pending)
      run.execute_now

      run.reload
      @install.reload

      assert_equal "succeeded", run.status
      assert_equal "ok", run.signal_status
      assert_equal "ok", @install.alert_state
      assert_equal 0, @install.consecutive_failures

      created_beep = @account.beeps.order(:created_at).last
      assert_equal "once", created_beep.kind
      assert_equal "Target Recovered", created_beep.title
      assert_equal @install.id, created_beep.beeper_install_id

      # Deliver the created once beep run
      beep_run = created_beep.runs.sole
      beep_run.deliver_now

      assert_equal 1, ActionMailer::Base.deliveries.size
      assert_equal "Target Recovered", ActionMailer::Base.deliveries.last.subject
    end
  end

  test "run execution routes to signals queue for beeper runs" do
    assert_enqueued_with(job: RunBeeperJob, queue: "signals") do
      run = @install.runs.create!(scheduled_for: Time.current, status: :pending)
      run.deliver_later
    end
  end
end
