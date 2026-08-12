require "test_helper"

class Account::InvitationAcceptancesControllerTest < ActionDispatch::IntegrationTest
  setup do
    @account = accounts(:john_account)
    @invitation = Account::Invitation.create!(
      account: @account,
      email: "newbie@example.com",
      invited_by: users(:john)
    )
  end

  test "accepting creates acceptance and keeps invitation" do
    identity = Identity.create!(email: "newbie@example.com")
    sign_in_as identity

    assert_difference -> { Account::InvitationAcceptance.count }, 1 do
      post account_invitation_acceptance_path(@invitation.token, script_name: @account.slug_path)
    end

    assert Account::Invitation.exists?(@invitation.id)
    assert_includes identity.reload.accounts, @account
    assert_redirected_to root_url(script_name: @account.slug_path)
    assert_equal @account.slug, identity.reload.last_account_slug
  end

  test "accepting with wrong email redirects with alert" do
    sign_in_as identities(:yuler)

    assert_no_difference -> { Account::InvitationAcceptance.count } do
      post account_invitation_acceptance_path(@invitation.token, script_name: @account.slug_path)
    end

    assert_redirected_to account_invitation_path(@invitation.token, script_name: @account.slug_path)
    assert_match(/email/i, flash[:alert])
  end

  test "accepting already responded invitation redirects with alert" do
    identity = Identity.create!(email: "newbie@example.com")
    @invitation.create_decline!(identity: identity)
    sign_in_as identity

    assert_no_difference -> { Account::InvitationAcceptance.count } do
      post account_invitation_acceptance_path(@invitation.token, script_name: @account.slug_path)
    end

    assert_redirected_to account_invitation_path(@invitation.token, script_name: @account.slug_path)
    assert_equal "This invitation has already been responded to.", flash[:alert]
  end
end
