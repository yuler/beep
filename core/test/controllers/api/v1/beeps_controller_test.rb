require "test_helper"

class Api::V1::BeepsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @run_at = 1.hour.from_now.change(usec: 0)
  end

  test "index returns account beeps newest first" do
    older = @account.beeps.create!(kind: :once, message: "Older", run_at: @run_at)
    newer = @account.beeps.create!(kind: :once, message: "Newer", run_at: @run_at + 1.hour)
    accounts(:yuler_account).beeps.create!(kind: :once, message: "Other", run_at: @run_at)

    get "/api/v1/#{@account.slug}/beeps",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    messages = response.parsed_body["beeps"].map { |beep| beep["message"] }
    assert_equal [ "Newer", "Older" ], messages
    assert_equal newer.id, response.parsed_body["beeps"].first["id"]
    assert_equal older.id, response.parsed_body["beeps"].second["id"]
  end

  test "index requires authentication" do
    get "/api/v1/#{@account.slug}/beeps", as: :json

    assert_response :unauthorized
  end

  test "index returns not found for another account" do
    get "/api/v1/#{accounts(:yuler_account).slug}/beeps",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
  end

  test "create makes a once beep and copies run_at to next_run_at" do
    assert_difference -> { @account.beeps.count }, 1 do
      post "/api/v1/#{@account.slug}/beeps",
        params: { message: "Call mom", run_at: @run_at.iso8601 },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :created
    body = response.parsed_body
    assert_equal "Call mom", body["message"]
    assert_equal "once", body["kind"]
    assert_equal "active", body["status"]
    assert_equal "UTC", body["timezone"]
    assert_equal @run_at.iso8601, Time.iso8601(body["run_at"]).iso8601
    assert_equal @run_at.iso8601, Time.iso8601(body["next_run_at"]).iso8601
  end

  test "create rejects a blank message" do
    post "/api/v1/#{@account.slug}/beeps",
      params: { message: "", run_at: @run_at.iso8601 },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert_equal "VALIDATION_ERROR", response.parsed_body["code"]
  end

  test "create rejects a run_at in the past" do
    post "/api/v1/#{@account.slug}/beeps",
      params: { message: "Call mom", run_at: 1.hour.ago.iso8601 },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert_equal "VALIDATION_ERROR", response.parsed_body["code"]
  end

  test "create requires authentication" do
    post "/api/v1/#{@account.slug}/beeps",
      params: { message: "Call mom", run_at: @run_at.iso8601 },
      as: :json

    assert_response :unauthorized
  end
end
