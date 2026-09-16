require "test_helper"

class Api::V1::Beepers::RunsControllerTest < ActionDispatch::IntegrationTest
  include ActiveJob::TestHelper

  setup do
    BeeperApp.seed_official
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @beeper_app = BeeperApp.find_by!(slug: "site-uptime")
    @beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "My Uptime",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )
  end

  test "index returns runs newest first" do
    3.times { |i| @beeper.runs.create!(scheduled_for: (i + 1).minutes.ago, status: :succeeded) }

    get "/api/v1/#{@account.slug}/beepers/#{@beeper.id}/runs",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    runs = response.parsed_body["runs"]
    assert_equal 3, runs.size
    scheduled = runs.map { |run| Time.zone.parse(run["scheduled_for"]) }
    assert_equal scheduled.sort.reverse, scheduled
  end

  test "index paginates runs with geared pagination" do
    20.times do |i|
      @beeper.runs.create!(scheduled_for: (i + 1).minutes.ago, status: :succeeded)
    end

    get "/api/v1/#{@account.slug}/beepers/#{@beeper.id}/runs",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    runs = response.parsed_body["runs"]
    pagination = response.parsed_body["pagination"]
    assert_equal 15, runs.size
    assert_equal true, pagination["has_more"]
    assert_not_nil pagination["next_page"]
    assert_equal 20, pagination["total_count"]
    newest = @beeper.runs.order(scheduled_for: :desc).first
    assert_equal newest.id, runs.first["id"]

    # Fetch second page
    get "/api/v1/#{@account.slug}/beepers/#{@beeper.id}/runs",
      params: { page: pagination["next_page"] },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    runs2 = response.parsed_body["runs"]
    pagination2 = response.parsed_body["pagination"]
    assert_equal 5, runs2.size
    assert_equal false, pagination2["has_more"]
    assert_nil pagination2["next_page"]
  end

  test "index returns empty runs when none exist" do
    get "/api/v1/#{@account.slug}/beepers/#{@beeper.id}/runs",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_empty response.parsed_body["runs"]
  end

  test "index returns not found for another account beeper" do
    other = Beeper.create!(
      account: accounts(:yuler_account),
      beeper_app: @beeper_app,
      title: "Other Account Beeper",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    get "/api/v1/#{@account.slug}/beepers/#{other.id}/runs",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
  end
end
