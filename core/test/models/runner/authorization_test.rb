require "test_helper"

class Runner::AuthorizationTest < ActiveSupport::TestCase
  setup do
    @user = users(:john)
  end

  test "create_request! creates authorization with device and user code" do
    auth = Runner::Authorization.create_request!(
      runner_name: "MacBook Pro",
      tags: %w[local macos],
      metadata: { "os" => "darwin", "arch" => "arm64" }
    )

    assert_predicate auth.device_code, :present?
    assert_predicate auth.user_code, :present?
    assert_match(/^[A-Z0-9]{4}-[A-Z0-9]{4}$/, auth.user_code)
    assert_equal "MacBook Pro", auth.runner_name
    assert_equal %w[local macos], auth.tags
    assert_equal({ "os" => "darwin", "arch" => "arm64" }, auth.metadata)
    assert_equal "pending", auth.status
    assert auth.expires_at > Time.current
  end

  test "approve! creates runner and updates authorization" do
    auth = Runner::Authorization.create_request!(runner_name: "MacBook Pro", tags: %w[local])

    assert_difference -> { Runner.count }, 1 do
      success = auth.approve!(user: @user, name: "Custom Runner", tags: %w[ci production])
      assert success
    end

    assert_equal "approved", auth.status
    assert_equal @user.account, auth.account
    assert_equal @user, auth.user
    assert_equal "Custom Runner", auth.runner.name
    assert_equal %w[ci production], auth.runner.tags
  end

  test "deny! marks authorization as access_denied" do
    auth = Runner::Authorization.create_request!(runner_name: "MacBook Pro")

    assert auth.deny!
    assert_equal "access_denied", auth.reload.status
  end

  test "consume_token! transitions approved to consumed and returns runner" do
    auth = Runner::Authorization.create_request!(runner_name: "MacBook Pro")
    auth.approve!(user: @user)

    runner = auth.consume_token!
    assert_not_nil runner
    assert_equal "consumed", auth.reload.status

    # Subsequent call returns nil
    assert_nil auth.consume_token!
  end
end
