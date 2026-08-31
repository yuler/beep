require "test_helper"

class BeeperTest < ActiveSupport::TestCase
  setup do
    BeeperApp.seed_official
    @account = accounts(:john_account)
    @beeper_app = BeeperApp.find_by!(slug: "site-uptime")
  end

  test "normalizes blank body to nil" do
    beeper = Beeper.new(
      account: @account,
      beeper_app: @beeper_app,
      title: "Example",
      body: "   ",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    assert beeper.valid?
    assert_nil beeper.body
  end

  test "validates body maximum length" do
    beeper = Beeper.new(
      account: @account,
      beeper_app: @beeper_app,
      title: "Example",
      body: "x" * (Beeper::BODY_MAX_LENGTH + 1),
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    assert_not beeper.valid?
    assert_includes beeper.errors[:body], "is too long (maximum is #{Beeper::BODY_MAX_LENGTH} characters)"
  end
end
