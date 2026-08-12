require "test_helper"

class Account::InvitationTest < ActiveSupport::TestCase
  setup do
    @account = accounts(:john_account)
    @invitation = Account::Invitation.create!(
      account: @account,
      email: "newbie@example.com",
      invited_by: users(:john)
    )
  end

  test "pending until a response exists" do
    assert @invitation.pending?
    assert_not @invitation.accepted?
    assert_not @invitation.declined?
  end

  test "accepted? when acceptance exists" do
    identity = Identity.create!(email: "newbie@example.com")
    user = identity.users.find_by!(account: accounts(:john_account)) rescue nil
    # join first so user exists for acceptance
    identity.join(@account, role: :member, verified_at: Time.current)
    user = identity.users.find_by!(account: @account)

    @invitation.create_acceptance!(identity: identity, user: user)

    assert @invitation.reload.accepted?
    assert_not @invitation.pending?
  end

  test "allows a new invitation after decline for same email" do
    identity = Identity.create!(email: "newbie@example.com")
    @invitation.create_decline!(identity: identity)

    assert_difference -> { Account::Invitation.count }, 1 do
      Account::Invitation.create!(
        account: @account,
        email: "newbie@example.com",
        invited_by: users(:john)
      )
    end
  end

  test "rejects a second pending invitation for same email" do
    duplicate = Account::Invitation.new(
      account: @account,
      email: "newbie@example.com",
      invited_by: users(:john)
    )

    assert_not duplicate.valid?
    assert_includes duplicate.errors[:email], "has already been invited"
  end

  test "accept! joins, records acceptance, keeps invitation" do
    identity = Identity.create!(email: "newbie@example.com")
    Current.identity = identity

    assert_difference -> { Account::InvitationAcceptance.count }, 1 do
      @invitation.accept!
    end

    assert Account::Invitation.exists?(@invitation.id)
    assert_includes identity.reload.accounts, @account
    assert @invitation.reload.accepted?
  end

  test "decline! records decline without joining" do
    identity = Identity.create!(email: "newbie@example.com")
    Current.identity = identity

    assert_difference -> { Account::InvitationDecline.count }, 1 do
      @invitation.decline!
    end

    assert Account::Invitation.exists?(@invitation.id)
    assert_not_includes identity.reload.accounts, @account
    assert @invitation.reload.declined?
  end

  test "accept! after decline raises AlreadyResponded" do
    identity = Identity.create!(email: "newbie@example.com")
    Current.identity = identity
    @invitation.decline!

    assert_raises(Account::Invitation::AlreadyResponded) { @invitation.accept! }
  end

  test "accept! with mismatched email raises EmailMismatch" do
    identity = Identity.create!(email: "other@example.com")
    Current.identity = identity

    assert_raises(Account::Invitation::EmailMismatch) { @invitation.accept! }
    assert_not_includes identity.reload.accounts, @account
  end

  test "decline! with mismatched email raises EmailMismatch" do
    identity = Identity.create!(email: "other@example.com")
    Current.identity = identity

    assert_raises(Account::Invitation::EmailMismatch) { @invitation.decline! }
    assert Account::Invitation.exists?(@invitation.id)
  end
end
