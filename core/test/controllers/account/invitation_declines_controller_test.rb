require "test_helper"

class Account::InvitationDeclinesControllerTest < ActionDispatch::IntegrationTest
  setup do
    @account = accounts(:john_account)
    @invitation = Account::Invitation.create!(
      account: @account,
      email: "newbie@example.com",
      invited_by: users(:john)
    )
  end

  test "declining creates decline and keeps invitation without joining" do
    identity = Identity.create!(email: "newbie@example.com")
    sign_in_as identity

    assert_difference -> { Account::InvitationDecline.count }, 1 do
      assert_no_difference -> { @account.users.count } do
        post account_invitation_decline_path(@invitation.token, script_name: @account.slug_path)
      end
    end

    assert Account::Invitation.exists?(@invitation.id)
    assert_not_includes identity.reload.accounts, @account
    assert_redirected_to my_accounts_url(script_name: nil)
  end

  test "declining with wrong email redirects with alert" do
    sign_in_as identities(:yuler)

    assert_no_difference -> { Account::InvitationDecline.count } do
      post account_invitation_decline_path(@invitation.token, script_name: @account.slug_path)
    end

    assert_redirected_to account_invitation_path(@invitation.token, script_name: @account.slug_path)
    assert_match(/email/i, flash[:alert])
  end

  test "declining already responded invitation redirects with alert" do
    identity = Identity.create!(email: "newbie@example.com")
    identity.join(@account, role: :member, verified_at: Time.current)
    user = identity.users.find_by!(account: @account)
    @invitation.create_acceptance!(identity: identity, user: user)
    sign_in_as identity

    assert_no_difference -> { Account::InvitationDecline.count } do
      post account_invitation_decline_path(@invitation.token, script_name: @account.slug_path)
    end

    assert_redirected_to account_invitation_path(@invitation.token, script_name: @account.slug_path)
    assert_equal "This invitation has already been responded to.", flash[:alert]
  end
end
