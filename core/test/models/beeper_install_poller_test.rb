require "test_helper"

class BeeperInstallPollerTest < ActiveSupport::TestCase
  setup do
    @account = accounts(:john_account)
    @beeper = Beeper.create!(slug: "echo", version: "1.0.0", manifest: echo_manifest)
  end

  test "poll_due_now claims an active install whose next_run_at is past" do
    install = BeeperInstall.create!(
      account: @account,
      beeper: @beeper,
      title: "Echo Install",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "status" => "ok" },
      notification_channels: %w[ email ]
    )
    install.update_columns(next_run_at: 1.minute.ago, status: "active")

    BeeperInstall.poll_due_now

    install.reload
    assert install.firing? || install.active?
    assert_equal 1, install.runs.count
    assert_equal "pending", install.runs.first.status
  end

  test "poll_due_now ignores paused and non-due installs" do
    paused_install = BeeperInstall.create!(
      account: @account,
      beeper: @beeper,
      title: "Paused Echo",
      cron: "*/5 * * * *",
      timezone: "UTC",
      status: "paused",
      next_run_at: 1.minute.ago,
      notification_channels: %w[ email ]
    )
    future_install = BeeperInstall.create!(
      account: @account,
      beeper: @beeper,
      title: "Future Echo",
      cron: "*/5 * * * *",
      timezone: "UTC",
      next_run_at: 5.minutes.from_now,
      notification_channels: %w[ email ]
    )

    BeeperInstall.poll_due_now

    assert_equal 0, paused_install.runs.count
    assert_equal 0, future_install.runs.count
  end

  test "pause! and resume! change install status and calculate next_run_at" do
    install = BeeperInstall.create!(
      account: @account,
      beeper: @beeper,
      title: "Echo Lifecycle",
      cron: "*/5 * * * *",
      timezone: "UTC",
      notification_channels: %w[ email ]
    )
    assert install.active?
    assert_not_nil install.next_run_at

    install.pause!
    assert install.paused?

    install.resume!
    assert install.active?
    assert_not_nil install.next_run_at
  end

  test "default notification_channels copies from account owner if empty" do
    install = BeeperInstall.create!(
      account: @account,
      beeper: @beeper,
      title: "Default Channels",
      cron: "*/5 * * * *",
      timezone: "UTC"
    )
    assert_equal @account.owner_user.notification_channels, install.notification_channels
  end

  private

  def echo_manifest
    {
      "manifest_version" => 1,
      "slug" => "echo",
      "name" => "Echo",
      "version" => "1.0.0",
      "author" => "Beep",
      "schedule" => { "default_cron" => "*/5 * * * *" }
    }
  end
end
