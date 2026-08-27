require "test_helper"

class BeeperPollerJobTest < ActiveSupport::TestCase
  include ActiveJob::TestHelper

  test "perform claims due beepers" do
    account = accounts(:john_account)
    beeper_app = BeeperApp.create!(
      slug: "echo",
      version: "1.0.0",
      manifest: {
        "manifest_version" => 1,
        "slug" => "echo",
        "name" => "Echo",
        "version" => "1.0.0",
        "author" => "Beep",
        "schedule" => { "default_cron" => "*/5 * * * *" }
      }
    )
    beeper = Beeper.create!(
      account: account,
      beeper_app: beeper_app,
      title: "Echo Signal",
      cron: "*/5 * * * *",
      timezone: "UTC",
      notification_channels: %w[ email ]
    )
    beeper.update_columns(next_run_at: 1.minute.ago)

    BeeperPollerJob.perform_now

    assert beeper.reload.firing? || beeper.active?
    assert_equal 1, beeper.runs.count
  end
end
