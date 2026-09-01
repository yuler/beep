require "test_helper"

class Api::V1::Runner::TasksControllerTest < ActionDispatch::IntegrationTest
  setup do
    BeeperApp.seed_official
    @account = accounts(:john_account)
    @beeper_app = BeeperApp.find_by!(slug: "site-uptime")
    @runner = @account.runners.create!(name: "Test-Runner", tags: ["intranet"])
    @runner_token = @runner.raw_token
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
    assert_equal "amd64", @runner.arch
    assert_equal "node-1", @runner.hostname
    assert_equal true, @runner.allow_exec
  end

  test "poll returns 204 when no tasks are due" do
    post "/api/v1/runner/tasks/poll",
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :no_content
  end

  test "poll claims pending task assigned to runner by runner_id" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      runner: @runner,
      title: "Intranet Uptime",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "http://192.168.1.10:8080/health" }
    )

    run = beeper.trigger_run!
    assert run.pending?
    assert_nil run.runner_id

    post "/api/v1/runner/tasks/poll",
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :success
    body = response.parsed_body
    task = body["task"]
    assert_equal run.id, task["id"]
    assert_equal beeper.id, task["beeper_id"]
    assert_equal "Intranet Uptime", task["title"]
    assert_equal "site-uptime", task["app_slug"]
    assert_equal "http://192.168.1.10:8080/health", task["config"]["target_url"]

    run.reload
    assert_equal "running", run.status
    assert_equal @runner.id, run.runner_id
    assert_not_nil run.claimed_at

    @runner.reload
    assert_equal "online", @runner.status
  end

  test "poll claims pending task matching runner tag" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      runner_tag: "intranet",
      title: "Tagged Intranet Uptime",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "http://192.168.1.20:8080/health" }
    )

    run = beeper.trigger_run!

    post "/api/v1/runner/tasks/poll",
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal run.id, body["task"]["id"]
  end

  test "result submits probe outcome and updates alert state and triggers notification if needed" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      runner: @runner,
      title: "Intranet Uptime",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "http://192.168.1.10:8080/health" }
    )

    run = beeper.trigger_run!
    run.update!(status: "running", runner: @runner, claimed_at: Time.current)

    # First alerting result: failure_threshold = 2, so state becomes pending
    post "/api/v1/runner/tasks/#{run.id}/result",
      params: {
        status: "alerting",
        title: "Connection Refused",
        message: "Failed to connect to 192.168.1.10:8080",
        metrics: { "status" => 0, "latency_ms" => 15 }
      },
      headers: { "X-Runner-Token" => @runner_token },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal "acknowledged", body["status"]
    assert_equal run.id, body["run_id"]
    assert_equal "alerting", body["signal_status"]

    run.reload
    assert_equal "succeeded", run.status
    assert_equal "alerting", run.signal_status
    assert_equal "Connection Refused", run.signal_result["title"]
    assert_equal 15, run.signal_result["metrics"]["latency_ms"]

    beeper.reload
    assert_equal "pending", beeper.alert_state
    assert_equal 1, beeper.consecutive_failures
    assert_equal "active", beeper.status
  end

  test "unauthorized request without token is rejected" do
    post "/api/v1/runner/ping", as: :json
    assert_response :unauthorized
  end
end
