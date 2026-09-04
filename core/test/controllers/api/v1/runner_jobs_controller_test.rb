require "test_helper"

class Api::V1::RunnerJobsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @runner = @account.runners.create!(name: "Office")
  end

  test "create list and show runner jobs" do
    post "/api/v1/#{@account.slug}/runners/#{@runner.id}/jobs",
      params: {
        name: "Intranet HTTP",
        slug: "intranet-http",
        cron: "*/5 * * * *",
        timeout_seconds: 20,
        config: { "target_url" => "http://10.0.0.5/health" }
      },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :created
    job = response.parsed_body
    assert_equal "intranet-http", job["slug"]
    assert_equal 20, job["timeout_seconds"]

    get "/api/v1/#{@account.slug}/runners/#{@runner.id}/jobs",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal 1, response.parsed_body["jobs"].size

    post "/api/v1/#{@account.slug}/runners/#{@runner.id}/jobs/#{job["id"]}/runs",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :created
    assert_equal "pending", response.parsed_body["status"]

    patch "/api/v1/#{@account.slug}/runners/#{@runner.id}/jobs/#{job["id"]}",
      params: {
        name: "Intranet HTTP Updated",
        slug: "intranet-http-updated",
        cron: "*/10 * * * *",
        timezone: "Asia/Shanghai",
        timeout_seconds: 45,
        description: "Updated health check job"
      },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    updated = response.parsed_body
    assert_equal "Intranet HTTP Updated", updated["name"]
    assert_equal "intranet-http-updated", updated["slug"]
    assert_equal "*/10 * * * *", updated["cron"]
    assert_equal "Asia/Shanghai", updated["timezone"]
    assert_equal 45, updated["timeout_seconds"]
    assert_equal "Updated health check job", updated.dig("config", "description")
  end
end
