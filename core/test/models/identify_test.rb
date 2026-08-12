require "test_helper"

class IdentityTest < ActiveSupport::TestCase
  test "downcases and strips email" do
    identity = Identity.new(email: " DOWNCASED@EXAMPLE.COM ")
    assert_equal("downcased@example.com", identity.email)
  end
end
