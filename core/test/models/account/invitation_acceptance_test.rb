require "test_helper"

class Account::InvitationAcceptanceTest < ActiveSupport::TestCase
  test "invitation_id is unique" do
    account = accounts(:john_account)
    invitation = Account::Invitation.create!(
      account: account, email: "a@example.com", invited_by: users(:john)
    )
    identity = Identity.create!(email: "a@example.com")
    identity.join(account, role: :member, verified_at: Time.current)
    user = identity.users.find_by!(account: account)

    Account::InvitationAcceptance.create!(invitation: invitation, identity: identity, user: user)

    duplicate = Account::InvitationAcceptance.new(
      invitation: invitation, identity: identity, user: user
    )
    assert_not duplicate.valid?
  end
end
