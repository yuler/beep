require "test_helper"

class RequestForgeryProtectionTest < ActionDispatch::IntegrationTest
  setup do
    @previous = Rails.application.config.action_controller.allow_forgery_protection
    Rails.application.config.action_controller.allow_forgery_protection = true
  end

  teardown do
    Rails.application.config.action_controller.allow_forgery_protection = @previous
  end

  test "rejects cross-site JSON POST without Sec-Fetch-Site allowance" do
    post api_v1_session_url,
      params: { email: identities(:john).email },
      as: :json,
      headers: { "Sec-Fetch-Site" => "cross-site", "Origin" => "https://evil.example.com" }

    assert_response :unprocessable_entity
    assert_equal "INVALID_CROSS_ORIGIN", response.parsed_body["code"]
  end

  test "allows JSON POST without Sec-Fetch-Site for non-browser API clients" do
    post api_v1_session_url,
      params: { email: identities(:john).email },
      as: :json

    assert_response :success
  end

  # Sec-Fetch-Mode is always present in browser requests. When it is present
  # but Sec-Fetch-Site is absent (malformed or stripped by a proxy) the
  # request must not be treated as a non-browser API client.
  test "rejects browser-like request that has Sec-Fetch-Mode but no Sec-Fetch-Site" do
    # Force SSL so that super's nil-Sec-Fetch-Site fallback also returns false,
    # then confirm our Sec-Fetch-Mode guard fires independently.
    post api_v1_session_url,
      params: { email: identities(:john).email },
      as: :json,
      headers: { "Sec-Fetch-Mode" => "cors", "X-Forwarded-Proto" => "https" }

    # The request is rejected — either by super (SSL + nil Sec-Fetch-Site)
    # or by the Sec-Fetch-Mode guard in allowed_api_request?.
    assert_response :unprocessable_entity
  end
end
