require "test_helper"

class RunnerTest < ActiveSupport::TestCase
  setup do
    @account = accounts(:john_account)
  end

  test "generates token on create with prefix and digest" do
    runner = @account.runners.create!(name: "HQ-Server")

    assert runner.persisted?
    assert runner.raw_token.start_with?("beep_rt_")
    assert_equal runner.token_prefix, runner.raw_token[0, 12]
    assert_equal Digest::SHA256.hexdigest(runner.raw_token), runner.token_digest
    assert_equal "offline", runner.status
  end

  test "finds runner by raw token" do
    runner = @account.runners.create!(name: "HQ-Server")
    found = Runner.find_by_raw_token(runner.raw_token)

    assert_equal runner.id, found.id
    assert_nil Runner.find_by_raw_token("invalid_token")
    assert_nil Runner.find_by_raw_token(nil)
    assert_nil Runner.find_by_raw_token("")
  end

  test "normalizes and cleans tags" do
    runner = @account.runners.create!(
      name: "HQ-Server",
      tags: [ " intranet ", "db", "intranet", "" ]
    )

    assert_equal [ "intranet", "db" ], runner.tags
  end

  test "matches tags correctly" do
    runner = @account.runners.create!(
      name: "HQ-Server",
      tags: [ "intranet", "internal-api" ]
    )

    assert runner.matches_tag?("intranet")
    assert runner.matches_tag?("internal-api")
    assert runner.matches_tag?(nil)
    assert runner.matches_tag?("")
    assert_not runner.matches_tag?("public")
  end

  test "touch_activity! updates metadata and status" do
    runner = @account.runners.create!(name: "HQ-Server")

    travel_to Time.zone.parse("2026-09-01 12:00:00 UTC") do
      runner.touch_activity!(
        version: "1.0.0",
        os: "linux",
        arch: "arm64",
        hostname: "nas-01",
        ip_address: "192.168.1.50",
        allow_exec: true,
        status: "idle"
      )

      runner.reload
      assert_equal "idle", runner.status
      assert_equal "1.0.0", runner.version
      assert_equal "linux", runner.os
      assert_equal "arm64", runner.arch
      assert_equal "nas-01", runner.hostname
      assert_equal "192.168.1.50", runner.ip_address
      assert_equal true, runner.allow_exec
      assert_equal Time.zone.parse("2026-09-01 12:00:00 UTC"), runner.last_seen_at
    end
  end

  test "mark_stale_offline! transitions inactive runners to offline" do
    active_runner = @account.runners.create!(name: "Active-Runner")
    active_runner.update_columns(status: "online", last_seen_at: 10.seconds.ago)

    stale_runner = @account.runners.create!(name: "Stale-Runner")
    stale_runner.update_columns(status: "idle", last_seen_at: 70.seconds.ago)

    never_seen_runner = @account.runners.create!(name: "Never-Seen")
    never_seen_runner.update_columns(status: "idle", last_seen_at: nil)

    Runner.mark_stale_offline!

    assert_equal "online", active_runner.reload.status
    assert_equal "offline", stale_runner.reload.status
    assert_equal "offline", never_seen_runner.reload.status
  end

  test "regenerates token on demand" do
    runner = @account.runners.create!(name: "HQ-Server")
    old_digest = runner.token_digest

    runner.regenerate_token!
    assert_not_equal old_digest, runner.token_digest
    assert runner.raw_token.start_with?("beep_rt_")
  end
end
