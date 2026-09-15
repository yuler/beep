require "test_helper"

class Cli::AuthorizationTest < ActiveSupport::TestCase
  setup do
    @identity = identities(:john)
  end

  test "create_request! creates authorization with device and user code" do
    auth = Cli::Authorization.create_request!(client_name: "John's Laptop")

    assert_predicate auth.device_code, :present?
    assert_predicate auth.user_code, :present?
    assert_match(/^[A-Z0-9]{4}-[A-Z0-9]{4}$/, auth.user_code)
    assert_equal "John's Laptop", auth.client_name
    assert_equal "pending", auth.status
    assert auth.expires_at > Time.current
  end

  test "approve! creates identity access token and updates authorization" do
    auth = Cli::Authorization.create_request!(client_name: "John's Laptop")

    assert_difference -> { @identity.access_tokens.count }, 1 do
      success = auth.approve!(identity: @identity, client_name: "Custom Laptop")
      assert success
    end

    assert_equal "approved", auth.status
    assert_equal @identity, auth.identity
    assert_not_nil auth.access_token
    assert_equal "Custom Laptop", auth.access_token.description
    assert_predicate auth.access_token.token, :present?
    assert auth.access_token.token.start_with?("beep_pat_")
  end

  test "deny! marks authorization as access_denied" do
    auth = Cli::Authorization.create_request!

    assert auth.deny!
    assert_equal "access_denied", auth.reload.status
  end

  test "consume_token! transitions approved to consumed and returns access token" do
    auth = Cli::Authorization.create_request!
    auth.approve!(identity: @identity)

    token = auth.consume_token!
    assert_not_nil token
    assert_equal "consumed", auth.reload.status

    # Subsequent call returns nil
    assert_nil auth.consume_token!
  end
end
