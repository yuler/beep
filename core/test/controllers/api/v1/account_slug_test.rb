require "test_helper"

class Api::V1::AccountSlugTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
  end

  test "path slug scopes the account" do
    get "/api/v1/#{@account.slug}/test/private",
      headers: { "Authorization" => "Bearer #{@token}" }

    assert_response :success
    assert_equal @account.slug, response.parsed_body["slug"]
  end

  test "header scopes the account when path has no slug" do
    get "/api/v1/test/private",
      headers: {
        "Authorization" => "Bearer #{@token}",
        "X-Account-Slug" => @account.slug
      }

    assert_response :success
    assert_equal @account.slug, response.parsed_body["slug"]
  end

  test "path slug wins over conflicting header" do
    get "/api/v1/#{@account.slug}/test/private",
      headers: {
        "Authorization" => "Bearer #{@token}",
        "X-Account-Slug" => accounts(:yuler_account).slug
      }

    assert_response :success
    assert_equal @account.slug, response.parsed_body["slug"]
  end

  test "unknown path slug returns 404 when authenticated" do
    get "/api/v1/no-such-acct/test/private",
      headers: { "Authorization" => "Bearer #{@token}" }

    assert_response :not_found
  end

  test "unknown path slug returns 401 when unauthenticated" do
    get "/api/v1/no-such-acct/test/private"

    assert_response :unauthorized
  end

  test "known path slug returns 401 when unauthenticated" do
    get "/api/v1/#{@account.slug}/test/private"

    assert_response :unauthorized
  end

  test "non-member path slug returns 404" do
    get "/api/v1/#{accounts(:yuler_account).slug}/test/private",
      headers: { "Authorization" => "Bearer #{@token}" }

    assert_response :not_found
  end

  test "no path or header falls back to personal account" do
    @account.update!(personal: true)

    get "/api/v1/test/private",
      headers: { "Authorization" => "Bearer #{@token}" }

    assert_response :success
    assert_equal @account.slug, response.parsed_body["slug"]
  end

  test "unknown header slug returns 404" do
    get "/api/v1/test/private",
      headers: {
        "Authorization" => "Bearer #{@token}",
        "X-Account-Slug" => "no-such-acct"
      }

    assert_response :not_found
  end
end
