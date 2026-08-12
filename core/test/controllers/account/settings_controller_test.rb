require "test_helper"

class Account::SettingsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @account = accounts(:john_account)
    sign_in_as identities(:john), account: @account
  end

  test "show settings" do
    get account_settings_path(script_name: @account.slug_path)

    assert_response :success
    assert_select "input[name=?]", "account[slug]"
    assert_match(/held for 30 days|cannot claim|no redirect/i, response.body)
  end

  test "update slug and redirect under new script name" do
    old_slug = @account.slug
    patch account_settings_path(script_name: @account.slug_path),
      params: { account: { name: @account.name, slug: "john_renamed", description: "Updated" } }

    assert_redirected_to account_settings_url(script_name: "/john_renamed")
    assert_equal "john_renamed", @account.reload.slug
    assert_nil Account.find_by(slug: old_slug)
    assert Account::SlugHold.active.exists?(slug: old_slug)
  end

  test "invalid slug re-renders with errors" do
    patch account_settings_path(script_name: @account.slug_path),
      params: { account: { name: @account.name, slug: "ab", description: "" } }

    assert_response :unprocessable_entity
  end
end
