require "test_helper"

class InviteCodeTest < ActiveSupport::TestCase
  test "generate creates a new invite code and returns its code" do
    assert_difference "InviteCode.count", +1 do
      assert_equal InviteCode.generate, InviteCode.last.code
    end
  end

  test "claim destroys the invite code" do
    code = InviteCode.generate

    assert_difference "InviteCode.count", -1 do
      InviteCode.claim code
    end
  end

  test "claim returns true if valid" do
    assert InviteCode.claim(InviteCode.generate)
  end

  test "claim is falsy if invalid" do
    assert_not InviteCode.claim("invalid")
  end
end
