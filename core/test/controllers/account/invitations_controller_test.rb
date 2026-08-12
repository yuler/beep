require "test_helper"

class Account::InvitationsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @account = accounts(:john_account)
    @invitation = Account::Invitation.create!(
      account: @account,
      email: "newbie@example.com",
      invited_by: users(:john)
    )
  end

  test "guest sees login and remembers return path" do
    get account_invitation_path(@invitation.token, script_name: @account.slug_path)

    assert_response :success
    assert_select "a", text: "Login"
    assert_select "button", text: "Accept invitation", count: 0
    assert_equal account_invitation_url(@invitation.token, script_name: @account.slug_path),
      session[:return_to_after_authenticating]
  end

  test "invitee sees accept and decline buttons" do
    identity = Identity.create!(email: "newbie@example.com")
    sign_in_as identity

    get account_invitation_path(@invitation.token, script_name: @account.slug_path)

    assert_response :success
    assert_select "form[action=?][method=post]",
      account_invitation_acceptance_path(@invitation.token, script_name: @account.slug_path)
    assert_select "button", text: "Accept invitation"
    assert_select "form[action=?][method=post]",
      account_invitation_decline_path(@invitation.token, script_name: @account.slug_path)
    assert_select "button", text: "Decline"
  end

  test "signed in with wrong email sees login prompt" do
    sign_in_as identities(:yuler)

    get account_invitation_path(@invitation.token, script_name: @account.slug_path)

    assert_response :success
    assert_select "button", text: "Accept invitation", count: 0
    assert_select "button", text: "Login"
  end

  test "wrong account slug returns not found" do
    get account_invitation_path(@invitation.token, script_name: accounts(:yuler_account).slug_path)

    assert_response :not_found
  end

  test "accepted invitation shows outcome without action buttons" do
    identity = Identity.create!(email: "newbie@example.com")
    identity.join(@account, role: :member, verified_at: Time.current)
    user = identity.users.find_by!(account: @account)
    @invitation.create_acceptance!(identity: identity, user: user)
    sign_in_as identity

    get account_invitation_path(@invitation.token, script_name: @account.slug_path)

    assert_response :success
    assert_select "button", text: "Accept invitation", count: 0
    assert_select "button", text: "Decline", count: 0
    assert_match(/joined/i, response.body)
  end

  test "declined invitation shows outcome without action buttons" do
    identity = Identity.create!(email: "newbie@example.com")
    @invitation.create_decline!(identity: identity)
    sign_in_as identity

    get account_invitation_path(@invitation.token, script_name: @account.slug_path)

    assert_response :success
    assert_select "button", text: "Accept invitation", count: 0
    assert_select "button", text: "Decline", count: 0
    assert_match(/declined/i, response.body)
  end

  test "index shows pending accepted and declined statuses" do
    accepted = Account::Invitation.create!(
      account: @account,
      email: "accepted@example.com",
      invited_by: users(:john)
    )
    declined = Account::Invitation.create!(
      account: @account,
      email: "declined@example.com",
      invited_by: users(:john)
    )
    identity = Identity.create!(email: "accepted@example.com")
    identity.join(@account, role: :member, verified_at: Time.current)
    accepted.create_acceptance!(identity: identity, user: identity.users.find_by!(account: @account))
    declined.create_decline!(identity: Identity.create!(email: "declined@example.com"))

    sign_in_as identities(:john), account: @account
    get account_invitations_path(script_name: @account.slug_path)

    assert_response :success
    assert_match(/Pending/, response.body)
    assert_match(/Accepted/, response.body)
    assert_match(/Declined/, response.body)
  end

  test "new invitation form" do
    sign_in_as identities(:john), account: @account

    get new_account_invitation_path(script_name: @account.slug_path)

    assert_response :success
    assert_select "form[action=?]", account_invitations_path(script_name: @account.slug_path)
    assert_select "input[name=?]", "account_invitation[email]"
  end

  test "users index invite member links to new invitation" do
    sign_in_as identities(:john), account: @account

    get account_users_path(script_name: @account.slug_path)

    assert_response :success
    assert_select "a.btn[href=?]", new_account_invitation_path(script_name: @account.slug_path), text: "Invite Member"
  end
end
