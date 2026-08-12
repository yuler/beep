require "test_helper"

class Account::JoinControllerTest < ActionDispatch::IntegrationTest
  setup do
    @team = Account.create_with_owner(
      account: { name: "Join Team", personal: false, slug: "join_team" },
      owner: { name: "John", identity: identities(:john) }
    )
    @code = @team.join_code.code
  end

  test "guest sees login instead of join and remembers return path" do
    get account_join_path(@code, script_name: @team.slug_path)

    assert_response :success
    assert_select "h2", text: "Join #{@team.name}"
    assert_select "a", text: "Login"
    assert_select "button", text: "Join account", count: 0
    assert_equal account_join_url(@code, script_name: @team.slug_path),
      session[:return_to_after_authenticating]
  end

  test "signed-in non-member sees join confirmation" do
    sign_in_as identities(:yuler)

    get account_join_path(@code, script_name: @team.slug_path)

    assert_response :success
    assert_match(/Signed in as #{Regexp.escape(identities(:yuler).email)}/, response.body)
    assert_select "button", text: "Join account"
    assert_select "a", text: "Login", count: 0
  end

  test "signed-in member sees continue instead of join" do
    sign_in_as identities(:john), account: @team

    get account_join_path(@code, script_name: @team.slug_path)

    assert_response :success
    assert_select "a", text: "Continue"
    assert_select "button", text: "Join account", count: 0
  end

  test "account-scoped join url 404s when slug and code account disagree" do
    sign_in_as identities(:yuler)

    get account_join_path(@code, script_name: accounts(:yuler_account).slug_path)

    assert_response :not_found
  end

  test "signed-in non-member can redeem join code" do
    sign_in_as identities(:yuler)

    post account_join_path(@code, script_name: @team.slug_path)

    assert_includes identities(:yuler).reload.accounts, @team
    assert_redirected_to landing_url(script_name: @team.slug_path)
  end

  test "guest cannot redeem join code without logging in" do
    assert_no_difference -> { @team.users.count } do
      post account_join_path(@code, script_name: @team.slug_path)
    end

    assert_redirected_to new_session_url(script_name: nil)
  end
end
