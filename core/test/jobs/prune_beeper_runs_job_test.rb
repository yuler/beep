require "test_helper"

class PruneBeeperRunsJobTest < ActiveJob::TestCase
  setup do
    BeeperApp.seed_official
    @account = accounts(:john_account)
    @beeper_app = BeeperApp.find_by!(slug: "site-uptime")
    @beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Prune Me",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )
  end

  test "deletes runs older than the retention period in batches" do
    stale = @beeper.runs.create!(scheduled_for: 31.days.ago, status: :succeeded)
    fresh = @beeper.runs.create!(scheduled_for: 1.day.ago, status: :succeeded)

    PruneBeeperRunsJob.perform_now

    assert_not BeeperRun.exists?(stale.id)
    assert BeeperRun.exists?(fresh.id)
  end

  test "keeps runs exactly within the retention period" do
    boundary = @beeper.runs.create!(scheduled_for: PruneBeeperRunsJob::RETENTION.ago + 1.minute, status: :succeeded)

    PruneBeeperRunsJob.perform_now

    assert BeeperRun.exists?(boundary.id)
  end
end
