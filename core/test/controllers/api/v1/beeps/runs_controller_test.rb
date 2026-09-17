require "test_helper"

class Api::V1::Beeps::RunsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @beep = @account.beeps.create!(
      kind: :recurring,
      cron: "0 * * * *",
      title: "Hourly Beep",
      timezone: "UTC"
    )
  end

  test "index returns runs newest first with pagination" do
    3.times { |i| @beep.runs.create!(scheduled_for: (i + 1).hours.ago, status: :succeeded) }

    get "/api/v1/#{@account.slug}/beeps/#{@beep.id}/runs",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    runs = response.parsed_body["runs"]
    pagination = response.parsed_body["pagination"]
    assert_equal 3, runs.size
    assert_equal 3, pagination["total_count"]
    assert_equal false, pagination["has_more"]
    assert_nil pagination["next_page"]
  end

  test "index paginates runs with geared pagination" do
    20.times do |i|
      @beep.runs.create!(scheduled_for: (i + 1).hours.ago, status: :succeeded)
    end

    get "/api/v1/#{@account.slug}/beeps/#{@beep.id}/runs",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal 15, body["runs"].size
    assert_equal true, body["pagination"]["has_more"]
    assert_not_nil body["pagination"]["next_page"]
    assert_equal 20, body["pagination"]["total_count"]
    assert_equal "20", response.headers["X-Total-Count"]
    assert_match(/rel="next"/, response.headers["Link"])

    # Page 2
    get "/api/v1/#{@account.slug}/beeps/#{@beep.id}/runs",
      params: { page: body["pagination"]["next_page"] },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body2 = response.parsed_body
    assert_equal 5, body2["runs"].size
    assert_equal false, body2["pagination"]["has_more"]
    assert_nil body2["pagination"]["next_page"]
  end

  test "create triggers a new run" do
    assert_difference -> { @beep.runs.count }, 1 do
      post "/api/v1/#{@account.slug}/beeps/#{@beep.id}/runs",
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :created
    assert_equal "pending", response.parsed_body["status"]
  end
end
