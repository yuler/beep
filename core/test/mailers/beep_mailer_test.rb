require "test_helper"

class BeepMailerTest < ActionMailer::TestCase
  test "beep formats subject with prefix and timestamp and includes card details" do
    account = accounts(:john_account)
    beep = Beep.create!(
      account: account,
      kind: :once,
      title: "Call mom",
      body: "Bring **milk**",
      run_at: 1.hour.from_now
    )
    scheduled_for = Time.utc(2026, 9, 20, 8, 25, 0)
    run = beep.runs.create!(scheduled_for: scheduled_for)

    user = users(:john)
    user.assign_timezone(name: "Asia/Shanghai")
    user.save!

    email = BeepMailer.beep(run, user: user)

    assert_equal [ "john@example.com" ], email.to
    assert_equal "[Beep] Call mom (09-20 16:25)", email.subject
    assert_match "Triggered at:", email.html_part.body.to_s
    assert_match "2026-09-20 16:25 CST", email.html_part.body.to_s
    assert_match beep.web_url, email.html_part.body.to_s
    assert_match "View in Beep", email.html_part.body.to_s
    assert_match "Bring milk", email.text_part.body.to_s
    assert_match "[Beep] Call mom", email.text_part.body.to_s
    assert_equal "List-Unsubscribe=One-Click", email["List-Unsubscribe-Post"].to_s
    assert_match %r{<http.+/email_channel_unsubscribes/}, email["List-Unsubscribe"].to_s
  end

  test "beep falls back to beep timezone or UTC when user timezone is blank" do
    account = accounts(:john_account)
    beep = Beep.create!(
      account: account,
      kind: :once,
      title: "Water plants",
      timezone: "UTC",
      run_at: 1.hour.from_now
    )
    scheduled_for = Time.utc(2026, 9, 20, 16, 25, 0)
    run = beep.runs.create!(scheduled_for: scheduled_for)

    user = users(:john)
    user.update!(timezone: nil)

    email = BeepMailer.beep(run, user: user)

    assert_equal "[Beep] Water plants (09-20 16:25)", email.subject
    assert_match "2026-09-20 16:25 UTC", email.html_part.body.to_s
  end
end
