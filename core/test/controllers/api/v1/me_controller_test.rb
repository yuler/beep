require "test_helper"

class Api::V1::MeControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @account = accounts(:john_account)
  end

  test "show returns identity and accounts" do
    get api_v1_me_url, headers: { "Authorization" => "Bearer #{@token}" }, as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal @identity.id, body["identity"]["id"]
    assert_equal @identity.email, body["identity"]["email"]
    assert_equal @identity.full_name, body["identity"]["name"]
    assert_equal @identity.staff?, body["identity"]["staff"]
    assert body["accounts"].any? { |account| account["slug"] == @account.slug }
    assert_nil body["last_account_slug"]
  end

  test "show returns last_account_slug from identity" do
    @identity.update!(last_account_slug: @account.slug)

    get api_v1_me_url, headers: { "Authorization" => "Bearer #{@token}" }, as: :json

    assert_response :success
    assert_equal @account.slug, response.parsed_body["last_account_slug"]
  end

  test "show accepts session cookie" do
    ActionDispatch::TestRequest.create.cookie_jar.tap do |cookie_jar|
      cookie_jar.signed[:session_id] = @session.signed_id
      cookies[:session_id] = cookie_jar[:session_id]
    end

    get api_v1_me_url, as: :json

    assert_response :success
    assert_equal @identity.id, response.parsed_body["identity"]["id"]
  end

  test "show requires authentication" do
    get api_v1_me_url, as: :json

    assert_response :unauthorized
  end

  test "update_last_account sets identity last_account_slug for a membership" do
    put api_v1_me_last_account_url,
      params: { slug: @account.slug },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal @account.slug, response.parsed_body["last_account_slug"]
    assert_equal @account.slug, @identity.reload.last_account_slug
  end

  test "update_last_account rejects unknown membership" do
    put api_v1_me_last_account_url,
      params: { slug: "no-such-acct" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
  end
end
