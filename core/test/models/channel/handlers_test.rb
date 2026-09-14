require "test_helper"

class Channel::HandlersTest < ActiveSupport::TestCase
  test "resolves correct handler class for each kind" do
    assert_equal Channel::Handlers::Cli, Channel::Handlers.for("cli")
    assert_equal Channel::Handlers::Email, Channel::Handlers.for("email")
    assert_equal Channel::Handlers::WebPush, Channel::Handlers.for("web_push")
    assert_equal Channel::Handlers::Webhook, Channel::Handlers.for("webhook")
  end

  test "raises NotImplementedError for unknown kind" do
    assert_raises(NotImplementedError) do
      Channel::Handlers.for("fax")
    end
  end

  test "channel instance resolves corresponding handler" do
    account = accounts(:john_account)
    user = users(:john)

    cli_channel = Channel.new(account: account, user: user, kind: :cli, name: "my-cli")
    assert_equal Channel::Handlers::Cli, cli_channel.handler

    email_channel = Channel.new(account: account, user: user, kind: :email, name: "my-email")
    assert_equal Channel::Handlers::Email, email_channel.handler

    push_channel = Channel.new(account: account, user: user, kind: :web_push, name: "my-push")
    assert_equal Channel::Handlers::WebPush, push_channel.handler
  end

  test "maintains backward compatibility aliases on Channel" do
    assert_equal Channel::Handlers::Cli, Channel::Cli
    assert_equal Channel::Handlers::Email, Channel::Email
    assert_equal Channel::Handlers::WebPush, Channel::WebPush
    assert_equal Channel::Handlers::Webhook, Channel::Webhook
  end
end
