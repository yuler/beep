require "test_helper"

class Api::V1::Runner::TasksControllerTest < ActionDispatch::IntegrationTest
  setup do
    @account = accounts(:john_account)
    @runner = @account.runners.create!(name: "Test-Runner")
    @runner.update_columns(status: "online", last_seen_at: 5.seconds.ago)
    @runner_token = @runner.raw_token
    @job = @runner.jobs.create!(
      name: "Intranet HTTP",
      slug: "intranet-http",
      cron: "*/5 * * * *",
      timezone: "UTC",
      timeout_seconds: 30,
      config: { "target_url" => "http://192.168.1.10/health" }
    )
  end

  test "ping endpoint updates runner activity" do
    post "/api/v1/runner/ping",
      params: {
        version: "0.1.0",
        os: "linux",
        arch: "amd64",
        hostname: "node-1",
        allow_exec: true
      },
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal "ok", body["status"]
    assert_equal @runner.id, body["runner_id"]

    @runner.reload
    assert_equal "0.1.0", @runner.version
    assert_equal "linux", @runner.os
    assert_equal true, @runner.allow_exec
  end

  test "poll returns 204 when no tasks are due" do
    post "/api/v1/runner/tasks/poll",
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :no_content
  end

  test "poll claims pending job run and exposes log and result urls" do
    run = @job.trigger_run!
    assert run.pending?

    post "/api/v1/runner/tasks/poll",
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :success
    task = response.parsed_body["task"]
    assert_equal run.id, task["id"]
    assert_equal "intranet-http", task["job_slug"]
    assert_equal "Intranet HTTP", task["name"]
    assert_equal "http://192.168.1.10/health", task["config"]["target_url"]
    assert_includes task["log_url"], "/api/v1/runner/tasks/#{run.id}/logs"
    assert_includes task["result_url"], "/api/v1/runner/tasks/#{run.id}/result"

    run.reload
    assert_equal "running", run.status
    assert_not_nil run.claimed_at
  end

  test "logs appends output then result records outcome" do
    run = @job.trigger_run!
    run.claim_for!(@runner)

    post "/api/v1/runner/tasks/#{run.id}/logs",
      params: { chunk: "checking health\n" },
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :success
    run.reload
    assert_equal "checking health\n", run.log

    post "/api/v1/runner/tasks/#{run.id}/result",
      params: {
        status: "ok",
        title: "healthy",
        message: "200 OK",
        metrics: { "latency_ms" => 12 }
      },
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :success
    run.reload
    assert_equal "succeeded", run.status
    assert_equal "ok", run.result_status
    assert_equal 12, run.result["metrics"]["latency_ms"]

    @job.reload
    assert @job.active?
    assert_not_nil @job.last_run_at
  end

  test "unauthorized request without token is rejected" do
    post "/api/v1/runner/ping", as: :json
    assert_response :unauthorized
  end
end
