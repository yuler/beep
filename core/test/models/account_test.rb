require "test_helper"

class AccountTest < ActiveSupport::TestCase
  test "slug_path is present for personal and team accounts" do
    personal = accounts(:john_account)
    team = Account.create_with_owner(
      account: { name: "Acme Team", personal: false, slug: "acme_team" },
      owner: { name: "John", identity: identities(:john) }
    )

    assert_equal "/john_account", personal.slug_path
    assert_equal "/acme_team", team.slug_path
  end

  test "personal accounts cannot be destroyed" do
    account = accounts(:john_account)

    assert_not account.destroy
    assert account.persisted?
  end

  test "rejects a second personal account for the same identity" do
    duplicate = Account.create_with_owner(
      account: { name: "Another Personal", personal: true, slug: "john_solo2" },
      owner: { name: "John", identity: identities(:john) }
    )

    assert_not duplicate.persisted?
    assert_includes duplicate.errors[:base], "Identity already has a personal account"
  end

  test "generates slug from name with random digits" do
    first = Account.create_with_owner(
      account: { name: "Tester", personal: false },
      owner: { name: "Tester", identity: identities(:john) }
    )
    second = Account.create_with_owner(
      account: { name: "Tester", personal: false },
      owner: { name: "Tester", identity: identities(:yuler) }
    )

    assert first.persisted?
    assert second.persisted?
    assert_match(/\Atester\d{4}\z/, first.slug)
    assert_match(/\Atester\d{4}\z/, second.slug)
    assert_not_equal first.slug, second.slug
  end

  test "renaming holds the previous slug so others cannot claim it" do
    account = accounts(:john_account)
    old_slug = account.slug

    account.update!(slug: "john_renamed")

    hold = Account::SlugHold.active.find_by(slug: old_slug)
    assert hold
    assert_equal account, hold.account
    assert Account.slug_taken?(old_slug)

    other = Account.create_with_owner(
      account: { name: "Squatter", personal: false, slug: old_slug },
      owner: { name: "Yuler", identity: identities(:yuler) }
    )

    assert_not other.persisted?
    assert_includes other.errors[:slug], "is unavailable"
  end

  test "account can reclaim its own held slug" do
    account = accounts(:john_account)
    old_slug = account.slug

    account.update!(slug: "john_renamed")
    assert account.update(slug: old_slug)
    assert_equal old_slug, account.reload.slug
    assert_not Account::SlugHold.active.exists?(slug: old_slug)
  end
end
