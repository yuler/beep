require "test_helper"

class CurrentTest < ActiveSupport::TestCase
  setup do
    Current.reset
  end

  teardown do
    Current.reset
  end

  test "setting session does not fill account from personal" do
    identity = identities(:john)
    accounts(:john_account).update!(personal: true)
    session = identity.sessions.create!

    Current.session = session

    assert_equal identity, Current.identity
    assert_nil Current.account
  end

  test "with_account preserves account when session is set afterward" do
    identity = identities(:john)
    account = accounts(:john_account)
    session = identity.sessions.create!

    Current.with_account(account) do
      Current.session = session
      assert_equal account, Current.account
      assert_equal users(:john), Current.user
    end
  end
end
