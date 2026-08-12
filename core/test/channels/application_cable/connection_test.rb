require "test_helper"

class ApplicationCable::ConnectionTest < ActionCable::Connection::TestCase
  tests ApplicationCable::Connection

  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
  end

  test "connects for member with valid account slug" do
    cookies.signed[:session_id] = @session.signed_id

    connect env: { "account_slug" => @account.slug }

    assert_equal users(:john), connection.current_user
  end

  test "rejects when account slug is unknown" do
    cookies.signed[:session_id] = @session.signed_id

    assert_reject_connection do
      connect env: { "account_slug" => "no-such-acct" }
    end
  end

  test "rejects when identity is not a member" do
    cookies.signed[:session_id] = @session.signed_id

    assert_reject_connection do
      connect env: { "account_slug" => accounts(:yuler_account).slug }
    end
  end
end
