require "test_helper"

class PluginCheckAndAlertTest < ActiveSupport::TestCase
  include ActiveJob::TestHelper
  setup do
    stub_web_push_dns_resolution
    @account = accounts(:john_account)
    @plugin = Plugin.create!(
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

    @beep = Beep.create!(
      account: @account,
      plugin: @plugin,
      kind: :recurring,
      cron: "*/5 * * * *",
      timezone: "UTC",
      title: "Echo Check Monitor",
      plugin_config: { "status" => "ok" }
    )

    ActionMailer::Base.deliveries.clear
  end

  test "healthy check stays silent and updates check_status to ok" do
    run = @beep.runs.create!(scheduled_for: Time.current, status: :pending)
    run.execute_check_now

    run.reload
    @beep.reload

    assert_equal "succeeded", run.status
    assert_equal "ok", run.check_status
    assert_equal "ok", @beep.alert_state
    assert_equal 0, @beep.consecutive_failures
    assert_empty ActionMailer::Base.deliveries
  end

  test "single failure below threshold does not notify but increments consecutive_failures" do
    @beep.update!(plugin_config: { "status" => "alerting", "title" => "Target Down", "message" => "HTTP 500" })
    run = @beep.runs.create!(scheduled_for: Time.current, status: :pending)
    run.execute_check_now

    run.reload
    @beep.reload

    assert_equal "succeeded", run.status
    assert_equal "alerting", run.check_status
    assert_equal "ok", @beep.alert_state
    assert_equal 1, @beep.consecutive_failures
    assert_empty ActionMailer::Base.deliveries
  end

  test "reaching failure threshold triggers alert notification and updates alert_state" do
    @beep.update!(
      consecutive_failures: 1,
      plugin_config: { "status" => "alerting", "title" => "Target Down", "message" => "HTTP 500" }
    )
    run = @beep.runs.create!(scheduled_for: Time.current, status: :pending)
    run.execute_check_now

    run.reload
    @beep.reload

    assert_equal "succeeded", run.status
    assert_equal "alerting", run.check_status
    assert_equal "alerting", @beep.alert_state
    assert_equal 2, @beep.consecutive_failures

    assert_equal 1, ActionMailer::Base.deliveries.size
    assert_equal "Target Down", ActionMailer::Base.deliveries.last.subject
  end

  test "subsequent failures while already alerting do not spam notifications" do
    @beep.update!(
      alert_state: :alerting,
      consecutive_failures: 2,
      plugin_config: { "status" => "alerting", "title" => "Target Down", "message" => "HTTP 500" }
    )
    run = @beep.runs.create!(scheduled_for: Time.current, status: :pending)
    run.execute_check_now

    run.reload
    @beep.reload

    assert_equal "succeeded", run.status
    assert_equal "alerting", @beep.alert_state
    assert_equal 3, @beep.consecutive_failures
    assert_empty ActionMailer::Base.deliveries
  end

  test "recovery from alerting sends recovery notification and resets alert_state" do
    @beep.update!(
      alert_state: :alerting,
      consecutive_failures: 3,
      plugin_config: { "status" => "ok", "title" => "Target Recovered", "message" => "Service back online" }
    )
    run = @beep.runs.create!(scheduled_for: Time.current, status: :pending)
    run.execute_check_now

    run.reload
    @beep.reload

    assert_equal "succeeded", run.status
    assert_equal "ok", run.check_status
    assert_equal "ok", @beep.alert_state
    assert_equal 0, @beep.consecutive_failures

    assert_equal 1, ActionMailer::Base.deliveries.size
    assert_equal "Target Recovered", ActionMailer::Base.deliveries.last.subject
  end

  test "run execution routes to checks queue for plugin beeps" do
    assert_enqueued_with(job: RunCheckJob, queue: "checks") do
      run = @beep.runs.create!(scheduled_for: Time.current, status: :pending)
      run.deliver_later
    end
  end
end
