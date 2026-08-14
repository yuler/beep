require "test_helper"

class Api::V1::WebPushControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @previous_public_key = Rails.application.config.x.vapid.public_key
    @previous_private_key = Rails.application.config.x.vapid.private_key
    Rails.application.config.x.vapid.public_key = "test-vapid-public"
    Rails.application.config.x.vapid.private_key = "test-vapid-private"
  end

  teardown do
    Rails.application.config.x.vapid.public_key = @previous_public_key
    Rails.application.config.x.vapid.private_key = @previous_private_key
  end

  test "show returns vapid public key" do
    get api_v1_web_push_url, headers: { "Authorization" => "Bearer #{@token}" }, as: :json

    assert_response :success
    assert_equal "test-vapid-public", response.parsed_body["vapid_public_key"]
  end

  test "show requires authentication" do
    get api_v1_web_push_url, as: :json

    assert_response :unauthorized
  end

  test "show rejects account scope" do
    get "/api/v1/#{@account.slug}/web_push",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
  end

  test "show returns unavailable when vapid public key is missing" do
    Rails.application.config.x.vapid.public_key = nil

    get api_v1_web_push_url, headers: { "Authorization" => "Bearer #{@token}" }, as: :json

    assert_response :service_unavailable
    assert_equal "WEB_PUSH_UNAVAILABLE", response.parsed_body["code"]
  end
end
