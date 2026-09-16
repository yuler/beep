require "test_helper"

class Api::V1::Beeps::StatsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
  end

  test "show returns beeps stats for the current account" do
    # Active beep due today
    @account.beeps.create!(kind: :once, title: "Today Active", run_at: Time.current.end_of_day - 1.hour)
    # Active beep due tomorrow
    @account.beeps.create!(kind: :once, title: "Tomorrow Active", run_at: 1.day.from_now)
    # Paused beep
    paused = @account.beeps.create!(kind: :once, title: "Paused", run_at: Time.current.end_of_day - 1.hour)
    paused.update_columns(status: "paused")

    get "/api/v1/#{@account.slug}/beeps/stats",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    stats = response.parsed_body["stats"]
    assert_equal 2, stats["active"]
    assert_equal 1, stats["due_today"]
    assert_equal 0, stats["firing"]
  end

  test "show requires authentication" do
    get "/api/v1/#{@account.slug}/beeps/stats", as: :json
    assert_response :unauthorized
  end
end
