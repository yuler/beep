require "test_helper"

class BeeperPollerJobTest < ActiveSupport::TestCase
  include ActiveJob::TestHelper

  test "perform claims due installs" do
    account = accounts(:john_account)
    beeper = Beeper.create!(
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
    install = BeeperInstall.create!(
      account: account,
      beeper: beeper,
      title: "Echo Signal",
      cron: "*/5 * * * *",
      timezone: "UTC",
      notification_channels: %w[ email ]
    )
    install.update_columns(next_run_at: 1.minute.ago)

    BeeperPollerJob.perform_now

    assert install.reload.firing? || install.active?
    assert_equal 1, install.runs.count
  end
end
