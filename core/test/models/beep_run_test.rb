require "test_helper"

class BeepRunTest < ActiveSupport::TestCase
  test "rejects a second run for the same beep and scheduled_for" do
    account = accounts(:john_account)
    beep = Beep.create!(
      account: account,
      kind: :once,
      title: "Call mom",
      run_at: 1.hour.from_now.change(usec: 0)
    )
    scheduled_for = beep.next_run_at
    beep.runs.create!(scheduled_for: scheduled_for, status: :pending)

    duplicate = beep.runs.new(scheduled_for: scheduled_for, status: :pending)
    assert_raises(ActiveRecord::RecordNotUnique) { duplicate.save(validate: false) }
  end

  test "deliver_to_channel builds structured payload with source, intent, and metadata" do
    account = accounts(:john_account)
    user = users(:john)
    channel = account.channels.create!(
      user: user,
      name: "My CLI",
      kind: "cli"
    )

    runner = account.runners.create!(name: "Mac Runner")
    job = runner.jobs.create!(
      account: account,
      name: "Offwork Notifier",
      slug: "offwork-notifier",
      cron: "0 18 * * *",
      timezone: "UTC"
    )

    beep = Beep.create!(
      account: account,
      kind: :once,
      title: "Ready to leave",
      body: "Work complete!",
      source: job,
      intent: "get_off_work",
      metadata: { "first_checkin_time" => "09:12:30" },
      notification_channels: [ channel.id ]
    )

    run = beep.runs.create!(scheduled_for: Time.current, status: :pending)
    result = run.send(:deliver_to_channel, channel)
    assert_equal "queued", result["status"]

    delivery = channel.deliveries.find(result["delivery_id"])
    assert_equal "beep.fired", delivery.payload["event"]
    assert_equal "runner_job", delivery.payload["source"]
    assert_equal "Runner::Job", delivery.payload["source_type"]
    assert_equal job.id, delivery.payload["source_id"]
    assert_equal "get_off_work", delivery.payload["intent"]
    assert_equal "Ready to leave", delivery.payload["title"]
    assert_equal "Work complete!", delivery.payload["body"]
    assert_equal "09:12:30", delivery.payload.dig("metadata", "first_checkin_time")
  end
end
