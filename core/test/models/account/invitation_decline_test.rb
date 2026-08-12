require "test_helper"

class Account::InvitationDeclineTest < ActiveSupport::TestCase
  test "invitation_id is unique" do
    account = accounts(:john_account)
    invitation = Account::Invitation.create!(
      account: account, email: "b@example.com", invited_by: users(:john)
    )
    identity = Identity.create!(email: "b@example.com")

    Account::InvitationDecline.create!(invitation: invitation, identity: identity)

    duplicate = Account::InvitationDecline.new(
      invitation: invitation, identity: identity
    )
    assert_not duplicate.valid?
  end
end
