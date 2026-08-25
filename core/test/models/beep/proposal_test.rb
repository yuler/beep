require "test_helper"
require "ostruct"

class Beep::ProposalTest < ActiveSupport::TestCase
  test "create maps a reminder into a confirmable proposal" do
    run_at = 1.hour.from_now.change(usec: 0)
    chat = fake_chat({
      "intent" => "create",
      "title" => "Call mom",
      "body" => "Bring milk",
      "run_at" => run_at.iso8601
    }.to_json)

    result = Beep::Proposal.create("明天打电话给妈", chat: chat)

    assert_equal "create", result.intent
    assert_equal "Call mom", result.title
    assert_equal "Bring milk", result.body
    assert_equal run_at, result.run_at
    assert_equal Beep::TIMEZONE, result.timezone
    assert_equal({}, result.errors)
    assert_nil result.message
    assert result.confirmable?
  end

  test "create marks a past run_at as not confirmable" do
    chat = fake_chat({
      "intent" => "create",
      "title" => "Call mom",
      "body" => nil,
      "run_at" => "2020-01-01T00:00:00Z"
    }.to_json)

    result = Beep::Proposal.create("一小时前叫我", chat: chat)

    assert_equal "must be in the future", result.errors["run_at"]
    assert_not result.confirmable?
  end

  test "create handles prompt without run_at as confirmable immediate proposal" do
    chat = fake_chat({
      "intent" => "create",
      "kind" => "once",
      "title" => "Buy groceries",
      "body" => nil,
      "run_at" => nil
    }.to_json)

    result = Beep::Proposal.create("提醒我买菜", chat: chat)

    assert_equal "create", result.intent
    assert_equal "once", result.kind
    assert_equal "Buy groceries", result.title
    assert_nil result.run_at
    assert_nil result.cron
    assert_equal({}, result.errors)
    assert result.confirmable?
  end

  test "create handles recurring cron schedule proposals" do
    chat = fake_chat({
      "intent" => "create",
      "kind" => "recurring",
      "title" => "看一下手机是否有电",
      "body" => nil,
      "cron" => "*/5 14-20 * * *"
    }.to_json)

    result = Beep::Proposal.create("每天下午两点到八点，每 5 分钟提醒我一下，看一下手机是否有电", chat: chat)

    assert_equal "create", result.intent
    assert_equal "recurring", result.kind
    assert_equal "看一下手机是否有电", result.title
    assert_nil result.run_at
    assert_equal "*/5 14-20 * * *", result.cron
    assert_equal({}, result.errors)
    assert result.confirmable?
  end

  test "create asks the user to rewrite when intent is other" do
    chat = fake_chat({
      "intent" => "other",
      "title" => nil,
      "body" => nil,
      "run_at" => nil
    }.to_json)

    result = Beep::Proposal.create("hello", chat: chat)

    assert_equal "other", result.intent
    assert_equal "Describe what to be reminded of.", result.message
    assert_not result.confirmable?
  end

  test "create raises when the model does not return JSON" do
    assert_raises Beep::Proposal::Error do
      Beep::Proposal.create("明天九点", chat: fake_chat("not json"))
    end
  end

  private
    def fake_chat(content)
      chat = Object.new
      chat.define_singleton_method(:with_instructions) { |_| chat }
      chat.define_singleton_method(:with_temperature) { |_| chat }
      chat.define_singleton_method(:with_params) { |**_| chat }
      chat.define_singleton_method(:ask) { |_| OpenStruct.new(content: content) }
      chat
    end
end
