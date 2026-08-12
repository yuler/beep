require "test_helper"

class IdentityPersonalAccountTest < ActiveSupport::TestCase
  test "creates personal account eagerly on identity create" do
    identity = Identity.create!(email: "newbie@example.com")

    personal = identity.accounts.personal.first
    assert personal.present?
    assert_match(/\Anewbie\d*\z/, personal.slug)
    assert_equal 1, identity.accounts.personal.count
  end

  test "personal_account returns the eager-created account" do
    identity = identities(:john)

    assert_equal accounts(:john_account), identity.personal_account
  end
end
