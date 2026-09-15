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

  # `Sec-Fetch-Mode` without `Sec-Fetch-Site` is what the production Nitro
  # `/api` proxy produces for legitimate non-browser clients (it injects
  # `Sec-Fetch-Mode: cors` and normalizes `Accept` to `*/*`). Browsers always
  # send `Sec-Fetch-Site`, so its absence means the request did not come from
  # a browser page and cannot be a cross-site forgery.
  test "allows proxied non-browser request with Sec-Fetch-Mode but no Sec-Fetch-Site" do
    # Force SSL so that super's nil-Sec-Fetch-Site fallback returns false,
    # then confirm allowed_api_request? permits it on its own.
    post api_v1_session_url,
      params: { email: identities(:john).email },
      as: :json,
      headers: {
        "Sec-Fetch-Mode" => "cors",
        "Accept" => "*/*",
        "X-Forwarded-Proto" => "https"
      }

    assert_response :success
  end
end
