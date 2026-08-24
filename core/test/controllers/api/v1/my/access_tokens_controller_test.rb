require "test_helper"

class Api::V1::My::AccessTokensControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @session = @identity.sessions.create!
    @token = @session.signed_id
  end

  test "index lists access tokens of current identity" do
    token1 = @identity.access_tokens.create!(description: "CLI Token", permission: "write")
    token2 = @identity.access_tokens.create!(description: "Read only", permission: "read")

    other_identity = identities(:yuler)
    token_other = other_identity.access_tokens.create!(description: "Yuler token", permission: "write")

    get api_v1_my_access_tokens_url, headers: { "Authorization" => "Bearer #{@token}" }, as: :json

    assert_response :success
    body = response.parsed_body
    token_ids = body["access_tokens"].map { |t| t["id"] }

    assert_includes token_ids, token1.id
    assert_includes token_ids, token2.id
    assert_not_includes token_ids, token_other.id
  end

  test "create creates a new access token" do
    assert_difference -> { @identity.access_tokens.count }, 1 do
      post api_v1_my_access_tokens_url,
        params: { access_token: { description: "Raycast Extension", permission: "write" } },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :created
    body = response.parsed_body
    assert_equal "Raycast Extension", body["access_token"]["description"]
    assert_equal "write", body["access_token"]["permission"]
    assert body["access_token"]["token"].present?
  end

  test "destroy deletes access token" do
    access_token = @identity.access_tokens.create!(description: "To delete", permission: "read")

    assert_difference -> { @identity.access_tokens.count }, -1 do
      delete api_v1_my_access_token_url(access_token),
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :no_content
  end

  test "api requests authenticate using Identity::AccessToken" do
    access_token = @identity.access_tokens.create!(description: "API Key", permission: "write")

    get api_v1_me_url, headers: { "Authorization" => "Bearer #{access_token.token}" }, as: :json
    assert_response :success
    assert_equal @identity.id, response.parsed_body["identity"]["id"]

    assert access_token.reload.last_used_at.present?
  end

  test "read-only access token blocks write requests" do
    access_token = @identity.access_tokens.create!(description: "Read-only Key", permission: "read")
    account = accounts(:john_account)

    get "/api/v1/#{account.slug}/beeps",
      headers: { "Authorization" => "Bearer #{access_token.token}" },
      as: :json
    assert_response :success

    post "/api/v1/#{account.slug}/beeps",
      params: { beep: { title: "Test Beep", due_at: 1.day.from_now.iso8601 } },
      headers: { "Authorization" => "Bearer #{access_token.token}" },
      as: :json
    assert_response :unauthorized
  end
end
