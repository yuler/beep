require "test_helper"

class LandingsControllerTest < ActionDispatch::IntegrationTest
  test "authenticated identity without account context redirects to account picker" do
    sign_in_as identities(:john)

    get root_url(script_name: nil)

    assert_redirected_to my_accounts_url(script_name: nil)
  end

  test "unauthenticated visit shows landing" do
    get root_url(script_name: nil)

    assert_response :success
  end
end
