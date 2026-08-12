require "test_helper"

class AccountSlugBoundaryTest < ActionDispatch::IntegrationTest
  test "long reserved prefixes are not truncated into fake slugs" do
    get "/invitations/some-token-value"

    # Reserved top-level segment — not treated as an account slug by middleware.
    assert_response :redirect
    assert_not_equal "/404", response.location.to_s
  end
end
