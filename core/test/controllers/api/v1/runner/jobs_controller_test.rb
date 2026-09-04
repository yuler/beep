require "test_helper"

class Api::V1::Runner::JobsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @account = accounts(:john_account)
    @runner = @account.runners.create!(name: "Test-Runner")
    @runner.update_columns(status: "online", last_seen_at: 5.seconds.ago)
    @runner_token = @runner.token
    @job = @runner.jobs.create!(
      name: "Existing Job",
      slug: "existing-job",
      cron: "*/5 * * * *",
      timezone: "UTC",
      timeout_seconds: 30
    )
  end

  test "index lists jobs for current runner" do
    get "/api/v1/runner/jobs",
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal 1, body["jobs"].size
    assert_equal "existing-job", body["jobs"].first["slug"]
  end

  test "create registers new runner job" do
    post "/api/v1/runner/jobs",
      params: {
        slug: "new-check",
        name: "New Check",
        cron: "* * * * *",
        timeout_seconds: 45
      },
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :created
    body = response.parsed_body
    assert_equal "new-check", body["job"]["slug"]
    assert_equal "New Check", body["job"]["name"]
    assert_equal "* * * * *", body["job"]["cron"]
    assert_equal 45, body["job"]["timeout_seconds"]

    assert @runner.jobs.exists?(slug: "new-check")
  end

  test "create upserts existing job" do
    post "/api/v1/runner/jobs",
      params: {
        slug: "existing-job",
        name: "Updated Name",
        cron: "0 * * * *"
      },
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :created
    @job.reload
    assert_equal "Updated Name", @job.name
    assert_equal "0 * * * *", @job.cron
  end

  test "sync upserts multiple jobs" do
    post "/api/v1/runner/jobs/sync",
      params: {
        jobs: [
          { slug: "job-one", name: "Job One", cron: "*/10 * * * *" },
          { slug: "job-two", name: "Job Two", cron: "0 0 * * *" }
        ]
      },
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal "ok", body["status"]
    assert_equal 2, body["synced_count"]
    assert @runner.jobs.exists?(slug: "job-one")
    assert @runner.jobs.exists?(slug: "job-two")
  end

  test "sync with id renames slug on existing job" do
    post "/api/v1/runner/jobs/sync",
      params: {
        jobs: [
          {
            id: @job.id,
            slug: "renamed-job",
            name: "Renamed Job",
            cron: "*/15 * * * *",
            timezone: "Asia/Shanghai"
          }
        ]
      },
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal 1, body["synced_count"]
    assert_equal @job.id, body["jobs"].first["id"]
    assert_equal "renamed-job", body["jobs"].first["slug"]

    @job.reload
    assert_equal "renamed-job", @job.slug
    assert_equal "Renamed Job", @job.name
    assert_equal "*/15 * * * *", @job.cron
    assert_equal "Asia/Shanghai", @job.timezone
    assert_equal 1, @runner.jobs.count
    assert_not @runner.jobs.exists?(slug: "existing-job")
  end

  test "destroy removes job by slug" do
    assert @runner.jobs.exists?(slug: "existing-job")

    delete "/api/v1/runner/jobs/existing-job",
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :no_content
    assert_not @runner.jobs.exists?(slug: "existing-job")
  end

  test "destroy returns not found for non-existent job" do
    delete "/api/v1/runner/jobs/non-existent",
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :not_found
  end
end
