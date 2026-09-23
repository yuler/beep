require "test_helper"

class Api::V1::BeepPreviewsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
  end

  test "create preview for instant beep" do
    post "/api/v1/#{@account.slug}/beep_preview",
      params: { title: "Instant Beep" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal true, body["valid"]
    assert_equal "once", body["kind"]
    assert_equal "Instant Beep", body["title"]
    assert_equal "Schedule", body["schedule_key"]
    assert_includes body["schedule_display"], "instant"
    assert_empty body["errors"]
  end

  test "create preview for delay beep" do
    post "/api/v1/#{@account.slug}/beep_preview",
      params: { title: "Tea Ready", in: "15m", timezone: "Asia/Shanghai" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal true, body["valid"]
    assert_equal "once", body["kind"]
    assert_equal "Delay", body["schedule_key"]
    assert_includes body["schedule_display"], "15m →"
    assert_includes body["schedule_display"], "Asia/Shanghai"
    assert_not_includes body["schedule_display"], "(in 15m"
    assert_not_nil body["run_at"]
    assert_empty body["errors"]
  end

  test "create preview for at beep" do
    post "/api/v1/#{@account.slug}/beep_preview",
      params: { title: "Doctor Appointment", at: "16:30", timezone: "Asia/Shanghai" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal true, body["valid"]
    assert_equal "once", body["kind"]
    assert_equal "Run At", body["schedule_key"]
    assert_includes body["schedule_display"], "16:30"
    assert_includes body["schedule_display"], "Asia/Shanghai"
    assert_not_nil body["run_at"]
    assert_empty body["errors"]
  end

  test "create preview for cron beep" do
    post "/api/v1/#{@account.slug}/beep_preview",
      params: { title: "Daily Standup", cron: "0 9 * * 1-5", timezone: "Asia/Shanghai" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal true, body["valid"]
    assert_equal "recurring", body["kind"]
    assert_equal "Cron", body["schedule_key"]
    assert_nil body["cron_description"]
    assert_includes body["schedule_display"], "0 9 * * 1-5"
    assert_includes body["schedule_display"], "Next:"
    assert_not_nil body["next_run_at"]
    assert_empty body["errors"]
  end

  test "create preview for invalid beep" do
    post "/api/v1/#{@account.slug}/beep_preview",
      params: { title: "", cron: "invalid cron" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal false, body["valid"]
    assert_not_empty body["errors"]
  end

  test "create preview requires authentication" do
    post "/api/v1/#{@account.slug}/beep_preview",
      params: { title: "Instant Beep" },
      as: :json

    assert_response :unauthorized
  end
end
