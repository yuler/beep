require "test_helper"

class BeepMailerTest < ActionMailer::TestCase
  test "reminder uses the beep title and unsubscribe headers" do
    account = accounts(:john_account)
    beep = Beep.create!(
      account: account,
      kind: :once,
      title: "Call mom",
      body: "Bring **milk**",
      run_at: 1.hour.from_now
    )
    run = beep.runs.create!(scheduled_for: Time.current)

    email = BeepMailer.reminder(run)

    assert_equal [ "john@example.com" ], email.to
    assert_equal "Call mom", email.subject
    assert_match beep.web_url, email.html_part.body.to_s
    assert_match "Bring milk", email.text_part.body.to_s
    assert_equal "List-Unsubscribe=One-Click", email["List-Unsubscribe-Post"].to_s
    assert_match %r{<http.+/email_channel_unsubscribes/}, email["List-Unsubscribe"].to_s
  end
end
