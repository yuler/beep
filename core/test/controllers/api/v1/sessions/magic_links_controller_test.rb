require "test_helper"

class Api::V1::Sessions::MagicLinksControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    post api_v1_session_url, params: { email: @identity.email }, as: :json
    @pending_token = response.parsed_body["pending_authentication_token"]
    @code = @identity.magic_links.order(:created_at).last.code
  end

  test "create exchanges code for session token" do
    post api_v1_session_magic_link_url, params: {
      code: @code,
      pending_authentication_token: @pending_token
    }, as: :json

    assert_response :success
    token = response.parsed_body["session_token"]
    assert token.present?
    assert Session.find_signed(token)
    assert cookies[:session_id].present?
  end

  test "create accepts pending token from cookie" do
    post api_v1_session_url, params: { email: @identity.email }, as: :json
    pending_cookie = cookies[:pending_authentication_token]
    assert pending_cookie.present?
    code = @identity.magic_links.order(:created_at).last.code

    post api_v1_session_magic_link_url, params: { code: code }, as: :json

    assert_response :success
    assert response.parsed_body["session_token"].present?
  end

  test "create rejects missing pending token" do
    open_session do |sess|
      sess.post sess.api_v1_session_magic_link_url, params: { code: @code }, as: :json
      sess.assert_response :unauthorized
    end
  end

  test "create rejects bad code" do
    post api_v1_session_magic_link_url, params: {
      code: "AAAAAA",
      pending_authentication_token: @pending_token
    }, as: :json

    assert_response :unauthorized
    assert_match(/Try another code/, response.parsed_body["message"])
  end
end
