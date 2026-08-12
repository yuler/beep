require "test_helper"

class Api::V1::SessionsControllerTest < ActionDispatch::IntegrationTest
  test "create sends magic link for existing identity" do
    identity = identities(:john)

    assert_difference -> { identity.magic_links.count }, 1 do
      post api_v1_session_url, params: { email: identity.email }, as: :json
    end

    assert_response :success
    body = response.parsed_body
    assert body["pending_authentication_token"].present?
    assert_nil body["code"]
    assert cookies[:pending_authentication_token].present?
  end

  test "create signs up a new identity" do
    email = "newbie@example.com"

    assert_difference -> { Identity.count }, 1 do
      post api_v1_session_url, params: { email: email }, as: :json
    end

    assert_response :success
    assert response.parsed_body["pending_authentication_token"].present?
  end

  test "create rejects invalid email" do
    post api_v1_session_url, params: { email: "not-an-email" }, as: :json

    assert_response :unprocessable_entity
    assert_equal "INVALID_EMAIL", response.parsed_body["code"]
  end

  test "destroy revokes the bearer session" do
    identity = identities(:john)
    session = identity.sessions.create!
    token = session.signed_id

    delete api_v1_session_url, headers: { "Authorization" => "Bearer #{token}" }, as: :json

    assert_response :success
    assert_nil Session.find_by(id: session.id)
  end

  test "destroy revokes cookie session" do
    identity = identities(:john)
    session = identity.sessions.create!
    sign_in_with_session_cookie(session)

    delete api_v1_session_url, as: :json

    assert_response :success
    assert_nil Session.find_by(id: session.id)
  end

  test "destroy requires authentication" do
    delete api_v1_session_url, as: :json

    assert_response :unauthorized
  end

  private
    def sign_in_with_session_cookie(session)
      ActionDispatch::TestRequest.create.cookie_jar.tap do |cookie_jar|
        cookie_jar.signed[:session_id] = session.signed_id
        cookies[:session_id] = cookie_jar[:session_id]
      end
    end
end
