require "test_helper"

class My::AccountsControllerTest < ActionDispatch::IntegrationTest
  test "index shows picker when identity has multiple accounts" do
    identity = identities(:john)
    Account.create_with_owner(
      account: { name: "John Team", personal: false, slug: "john_team" },
      owner: { name: "John", identity: identity }
    )

    sign_in_as identity

    get my_accounts_url(script_name: nil)

    assert_response :success
    assert_select "h1", text: "Choose Your Account"
  end

  test "index shows picker when identity has exactly one account" do
    identity = identities(:john)
    assert_equal 1, identity.accounts.count

    sign_in_as identity

    get my_accounts_url(script_name: nil)

    assert_response :success
    assert_select "h1", text: "Choose Your Account"
    assert_select "a", text: "Create a New Account"
  end
end
