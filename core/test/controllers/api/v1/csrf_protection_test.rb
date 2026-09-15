require "test_helper"

class Api::V1::CsrfProtectionTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @previous_forgery_protection = Api::V1::BaseController.allow_forgery_protection
    Api::V1::BaseController.allow_forgery_protection = true
  end

  teardown do
    Api::V1::BaseController.allow_forgery_protection = @previous_forgery_protection
  end

  test "cookie POST without XHR header is rejected" do
    sign_in_with_cookie

    post "/api/v1/channels",
      params: { channel: { name: "laptop", kind: "cli" } },
      as: :json

    assert_response :unprocessable_entity
    assert_equal "INVALID_CROSS_ORIGIN", response.parsed_body["code"]
  end

  test "cookie POST with XHR header succeeds" do
    sign_in_with_cookie

    post "/api/v1/channels",
      params: { channel: { name: "laptop", kind: "cli" } },
      headers: { "X-Requested-With" => "XMLHttpRequest" },
      as: :json

    assert_response :created
  end

  test "cookie GET without XHR header succeeds" do
    sign_in_with_cookie

    get "/api/v1/channels", as: :json

    assert_response :success
  end

  test "bearer POST without XHR header succeeds" do
    post "/api/v1/channels",
      params: { channel: { name: "laptop", kind: "cli" } },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :created
  end

  private
    def sign_in_with_cookie
      ActionDispatch::TestRequest.create.cookie_jar.tap do |cookie_jar|
        cookie_jar.signed[:session_id] = @session.signed_id
        cookies[:session_id] = cookie_jar[:session_id]
      end
    end
end
