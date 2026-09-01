require "test_helper"

class SessionsControllerTest < ActionDispatch::IntegrationTest
  test "new renders the responsive auth screen" do
    get new_session_path

    assert_response :success
    assert_select "section.auth-stage[aria-labelledby=auth-title]"
    assert_select "h1#auth-title", text: "Sign in to your workspace."
    assert_select ".auth-brand__mark svg[viewBox='0 0 512 512']"
    assert_select ".field__control .field__icon svg[stroke='currentColor']"
    assert_select "input[type=email][name=email][autocomplete=username]"
    assert_select "footer", count: 0
  end

  test "sign in with one account lands on that account" do
    identity = identities(:john)
    assert_equal 1, identity.accounts.count

    post session_url(script_name: nil), params: { email: identity.email }
    magic_link = identity.magic_links.order(:created_at).last

    post session_magic_link_url(script_name: nil), params: { code: magic_link.code }

    assert_redirected_to root_url(script_name: accounts(:john_account).slug_path)
  end

  test "sign in with multiple accounts lands on account picker" do
    identity = identities(:john)
    Account.create_with_owner(
      account: { name: "John Team", personal: false, slug: "john_team" },
      owner: { name: "John", identity: identity }
    )

    post session_url(script_name: nil), params: { email: identity.email }
    magic_link = identity.magic_links.order(:created_at).last

    post session_magic_link_url(script_name: nil), params: { code: magic_link.code }

    assert_redirected_to my_accounts_url(script_name: nil)
  end

  test "logout with return_to sends user to login for that path" do
    sign_in_as identities(:john)
    return_to = account_invitation_path("invite-token", script_name: "/john_account")

    delete session_url(script_name: nil), params: {
      return_to: return_to,
      email: "newbie@example.com"
    }

    assert_redirected_to new_session_path(script_name: nil, email: "newbie@example.com")
    assert_equal return_to, session[:return_to_after_authenticating]
  end

  test "logout ignores absolute return_to urls" do
    sign_in_as identities(:john)

    delete session_url(script_name: nil), params: {
      return_to: "https://evil.example/phish",
      email: "newbie@example.com"
    }

    assert_redirected_to root_path
    assert_nil session[:return_to_after_authenticating]
  end
end
