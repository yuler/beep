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
    older = @account.beeps.create!(kind: :once, title: "Older", run_at: @run_at)
    newer = @account.beeps.create!(kind: :once, title: "Newer", run_at: @run_at + 1.hour)
    accounts(:yuler_account).beeps.create!(kind: :once, title: "Other", run_at: @run_at)

    get "/api/v1/#{@account.slug}/beeps",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    titles = response.parsed_body["beeps"].map { |beep| beep["title"] }
    assert_equal [ "Newer", "Older" ], titles
    assert_equal newer.id, response.parsed_body["beeps"].first["id"]
    assert_equal older.id, response.parsed_body["beeps"].second["id"]
    assert_equal [], response.parsed_body["beeps"].first["runs"]
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
        params: { title: "Call mom", body: "Bring **milk**", run_at: @run_at.iso8601 },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :created
    body = response.parsed_body
    assert_equal "Call mom", body["title"]
    assert_equal "Bring **milk**", body["body"]
    assert_equal "once", body["kind"]
    assert_nil body["cron"]
    assert_equal "active", body["status"]
    assert_equal Beep::TIMEZONE, body["timezone"]
    assert_equal @run_at.iso8601, Time.iso8601(body["run_at"]).iso8601
    assert_equal @run_at.iso8601, Time.iso8601(body["next_run_at"]).iso8601
    assert_nil body["channels"]
  end

  test "create without run_at triggers immediately as an ingest" do
    assert_enqueued_with(job: DeliverBeepRunJob) do
      assert_difference -> { @account.beeps.count }, 1 do
        post "/api/v1/#{@account.slug}/beeps",
          params: { title: "Alert right now" },
          headers: { "Authorization" => "Bearer #{@token}" },
          as: :json
      end
    end

    assert_response :created
    body = response.parsed_body
    assert_equal "Alert right now", body["title"]
    assert_equal "once", body["kind"]
    assert_equal "firing", body["status"]
    assert_not_nil body["run_at"]
    assert_equal 1, body["runs"].size
    assert_equal "pending", body["runs"].first["status"]
  end

  test "runs create triggers a new run for an existing beep" do
    beep = @account.beeps.create!(kind: :once, title: "Test Trigger", run_at: @run_at)

    assert_enqueued_with(job: DeliverBeepRunJob) do
      assert_difference -> { beep.runs.count }, 1 do
        post "/api/v1/#{@account.slug}/beeps/#{beep.id}/runs",
          headers: { "Authorization" => "Bearer #{@token}" },
          as: :json
      end
    end

    assert_response :created
    body = response.parsed_body
    assert_equal "pending", body["status"]
    assert beep.reload.firing?
  end

  test "create rejects a blank title" do
    post "/api/v1/#{@account.slug}/beeps",
      params: { title: "", run_at: @run_at.iso8601 },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert_equal "VALIDATION_ERROR", response.parsed_body["code"]
  end

  test "create requires authentication" do
    post "/api/v1/#{@account.slug}/beeps",
      params: { title: "Call mom", run_at: @run_at.iso8601 },
      as: :json

    assert_response :unauthorized
  end

  test "show returns the account beep" do
    beep = @account.beeps.create!(kind: :once, title: "Call mom", run_at: @run_at)

    get "/api/v1/#{@account.slug}/beeps/#{beep.id}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal beep.id, body["id"]
    assert_equal "Call mom", body["title"]
    assert_nil body["body"]
    assert_equal "once", body["kind"]
    assert_nil body["cron"]
    assert_equal "active", body["status"]
    assert_equal [], body["runs"]
    assert_nil body["channels"]
  end

  test "show includes beep runs newest first" do
    beep = @account.beeps.create!(kind: :once, title: "Call mom", run_at: @run_at)
    older = beep.runs.create!(scheduled_for: 2.hours.ago.change(usec: 0), status: :expired)
    newer = beep.runs.create!(
      scheduled_for: 1.hour.ago.change(usec: 0),
      status: :succeeded,
      result: { "web_push" => { "reason" => "no_subscriptions" } }
    )

    get "/api/v1/#{@account.slug}/beeps/#{beep.id}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    runs = response.parsed_body["runs"]
    assert_equal [ newer.id, older.id ], runs.map { |run| run["id"] }
    assert_equal "succeeded", runs.first["status"]
    assert_equal({ "web_push" => { "reason" => "no_subscriptions" } }, runs.first["result"])
    assert_equal newer.scheduled_for.iso8601, Time.iso8601(runs.first["scheduled_for"]).iso8601
  end

  test "show requires authentication" do
    beep = @account.beeps.create!(kind: :once, title: "Call mom", run_at: @run_at)

    get "/api/v1/#{@account.slug}/beeps/#{beep.id}", as: :json

    assert_response :unauthorized
  end

  test "show returns not found for another account's beep" do
    beep = accounts(:yuler_account).beeps.create!(kind: :once, title: "Other", run_at: @run_at)

    get "/api/v1/#{@account.slug}/beeps/#{beep.id}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
  end

  test "update changes the title" do
    beep = @account.beeps.create!(kind: :once, title: "Call mom", run_at: @run_at)

    patch "/api/v1/#{@account.slug}/beeps/#{beep.id}",
      params: { title: "Call dad" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "Call dad", response.parsed_body["title"]
    assert_equal "Call dad", beep.reload.title
  end

  test "destroy deletes the beep and its runs" do
    beep = @account.beeps.create!(kind: :once, title: "Call mom", run_at: @run_at)
    beep.runs.create!(scheduled_for: @run_at, status: :succeeded)

    assert_difference -> { Beep.count }, -1 do
      assert_difference -> { BeepRun.count }, -1 do
        delete "/api/v1/#{@account.slug}/beeps/#{beep.id}",
          headers: { "Authorization" => "Bearer #{@token}" },
          as: :json
      end
    end

    assert_response :no_content
    assert_nil Beep.find_by(id: beep.id)
  end

  test "destroy requires authentication" do
    beep = @account.beeps.create!(kind: :once, title: "Call mom", run_at: @run_at)

    delete "/api/v1/#{@account.slug}/beeps/#{beep.id}", as: :json

    assert_response :unauthorized
  end

  test "destroy returns not found for another account's beep" do
    beep = accounts(:yuler_account).beeps.create!(kind: :once, title: "Other", run_at: @run_at)

    delete "/api/v1/#{@account.slug}/beeps/#{beep.id}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
    assert beep.reload.present?
  end
end
