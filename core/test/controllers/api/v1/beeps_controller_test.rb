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

    pagination = response.parsed_body["pagination"]
    assert_equal 1, pagination["page"]
    assert_equal 2, pagination["total_count"]
    assert_equal false, pagination["has_more"]
    assert_nil pagination["next_page"]
    assert_equal "2", response.headers["X-Total-Count"]
  end

  test "index sorts by title and created_at" do
    older = @account.beeps.create!(kind: :once, title: "Bravo", run_at: @run_at)
    newer = @account.beeps.create!(kind: :once, title: "Alpha", run_at: @run_at + 1.hour)

    get "/api/v1/#{@account.slug}/beeps",
      params: { sort: "title", dir: "asc" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal [ "Alpha", "Bravo" ], response.parsed_body["beeps"].map { |beep| beep["title"] }

    get "/api/v1/#{@account.slug}/beeps",
      params: { sort: "created_at", dir: "asc" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal [ older.id, newer.id ], response.parsed_body["beeps"].map { |beep| beep["id"] }

    get "/api/v1/#{@account.slug}/beeps",
      params: { sort: "not_a_column", dir: "asc" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal [ newer.id, older.id ], response.parsed_body["beeps"].map { |beep| beep["id"] }
  end

  test "index filters by search query q across title, body, and id" do
    match_title = @account.beeps.create!(kind: :once, title: "Special Invoice reminder", run_at: @run_at)
    match_body = @account.beeps.create!(kind: :once, title: "Meeting", body: "Check invoice details", run_at: @run_at)
    other = @account.beeps.create!(kind: :once, title: "Workout", body: "Gym session", run_at: @run_at)

    get "/api/v1/#{@account.slug}/beeps",
      params: { q: "invoice" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    ids = response.parsed_body["beeps"].map { |b| b["id"] }
    assert_includes ids, match_title.id
    assert_includes ids, match_body.id
    assert_not_includes ids, other.id

    # Search by ID
    get "/api/v1/#{@account.slug}/beeps",
      params: { q: other.id },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    ids = response.parsed_body["beeps"].map { |b| b["id"] }
    assert_equal [ other.id ], ids

    # Search by non-UUID text does not crash on UUID column lookups
    get "/api/v1/#{@account.slug}/beeps",
      params: { q: "plain-text-not-a-uuid" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal [], response.parsed_body["beeps"]
  end

  test "index paginates with geared cursor and sets Link headers" do
    18.times do |i|
      @account.beeps.create!(kind: :once, title: "Beep #{i}", run_at: @run_at + i.minutes)
    end

    get "/api/v1/#{@account.slug}/beeps",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal 15, body["beeps"].size
    assert_equal true, body["pagination"]["has_more"]
    assert_not_nil body["pagination"]["next_page"]
    assert_equal 18, body["pagination"]["total_count"]
    assert_equal "18", response.headers["X-Total-Count"]
    assert_match(/rel="next"/, response.headers["Link"])

    # Fetch second page using next_page cursor
    next_cursor = body["pagination"]["next_page"]
    get "/api/v1/#{@account.slug}/beeps",
      params: { page: next_cursor },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body2 = response.parsed_body
    assert_equal 3, body2["beeps"].size
    assert_equal false, body2["pagination"]["has_more"]
    assert_nil body2["pagination"]["next_page"]
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
        params: { title: "Call mom", body: "Bring **milk**", run_at: @run_at.iso8601, intent: "lunch_break" },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :created
    body = response.parsed_body
    assert_equal "Call mom", body["title"]
    assert_equal "Bring **milk**", body["body"]
    assert_equal "once", body["kind"]
    assert_equal "lunch_break", body["intent"]
    assert_nil body["cron"]
    assert_equal "active", body["status"]
    assert_equal "UTC", body["timezone"]
    assert_equal @run_at.iso8601, Time.iso8601(body["run_at"]).iso8601
    assert_equal @run_at.iso8601, Time.iso8601(body["next_run_at"]).iso8601
    assert_equal [ "email", "web_push" ], body["notification_channels"]
  end

  test "create without run_at triggers immediately" do
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
    assert_includes response.parsed_body["errors"], "Title can't be blank"
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
    assert_equal [ "email", "web_push" ], body["notification_channels"]
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
      assert_difference -> { Beep::Run.count }, -1 do
        delete "/api/v1/#{@account.slug}/beeps/#{beep.id}",
          headers: { "Authorization" => "Bearer #{@token}" },
          as: :json
      end
    end

    assert_response :no_content
    assert_nil Beep.find_by(id: beep.id)
  end

  test "destroy deletes the beep with channel deliveries" do
    beep = @account.beeps.create!(kind: :once, title: "Call mom", run_at: @run_at)
    run = beep.runs.create!(scheduled_for: @run_at, status: :succeeded)
    channel = @account.channels.create!(user: users(:john), kind: :cli, name: "laptop")
    delivery = channel.deliveries.create!(beep_run: run, payload: { title: "Test" })

    assert_difference -> { Beep.count }, -1 do
      assert_difference -> { Beep::Run.count }, -1 do
        assert_difference -> { Channel::Delivery.count }, -1 do
          delete "/api/v1/#{@account.slug}/beeps/#{beep.id}",
            headers: { "Authorization" => "Bearer #{@token}" },
            as: :json
        end
      end
    end

    assert_response :no_content
    assert_nil Beep.find_by(id: beep.id)
    assert_nil Channel::Delivery.find_by(id: delivery.id)
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

  test "create recurring beep validates cron and calculates next_run_at" do
    post "/api/v1/#{@account.slug}/beeps",
      params: { kind: "recurring", title: "Daily meeting", cron: "0 9 * * *" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :created
    body = response.parsed_body
    assert_equal "recurring", body["kind"]
    assert_equal "0 9 * * *", body["cron"]
    assert_not_nil body["next_run_at"]
  end

  test "pause and resume updates beep status via pauses resource" do
    beep = @account.beeps.create!(kind: :recurring, title: "Standup", cron: "0 9 * * *")

    post "/api/v1/#{@account.slug}/beeps/#{beep.id}/pause",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "paused", response.parsed_body["status"]
    assert beep.reload.paused?

    delete "/api/v1/#{@account.slug}/beeps/#{beep.id}/pause",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "active", response.parsed_body["status"]
    assert beep.reload.active?
  end

  test "pause rejects a completed beep" do
    beep = @account.beeps.create!(kind: :once, title: "Done", run_at: @run_at)
    beep.update_columns(status: "completed", next_run_at: nil)

    post "/api/v1/#{@account.slug}/beeps/#{beep.id}/pause",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert_equal "VALIDATION_ERROR", response.parsed_body["code"]
    assert beep.reload.completed?
  end

  test "create uses the current user's timezone" do
    users(:john).update!(timezone: "Europe/London", timezone_source: "manual")

    post "/api/v1/#{@account.slug}/beeps",
      params: { title: "Call mom", run_at: @run_at.iso8601, timezone: "Asia/Tokyo" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :created
    assert_equal "Europe/London", response.parsed_body["timezone"]
  end

  test "create falls back to request timezone then UTC" do
    post "/api/v1/#{@account.slug}/beeps",
      params: { title: "Call mom", run_at: @run_at.iso8601, timezone: "Asia/Tokyo" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :created
    assert_equal "Asia/Tokyo", response.parsed_body["timezone"]

    post "/api/v1/#{@account.slug}/beeps",
      params: { title: "Call dad", run_at: @run_at.iso8601 },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :created
    assert_equal "UTC", response.parsed_body["timezone"]
  end

  test "update ignores a timezone in the payload" do
    beep = @account.beeps.create!(kind: :once, title: "Call mom", run_at: @run_at, timezone: "UTC")

    patch "/api/v1/#{@account.slug}/beeps/#{beep.id}",
      params: { title: "Call dad", timezone: "Asia/Tokyo" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "Call dad", response.parsed_body["title"]
    assert_equal "UTC", response.parsed_body["timezone"]
    assert_equal "UTC", beep.reload.timezone
  end

  test "index filters by status" do
    active_beep = @account.beeps.create!(kind: :once, title: "Active Beep", run_at: @run_at)
    firing_beep = @account.beeps.create!(kind: :once, title: "Firing Beep", run_at: @run_at)
    firing_beep.update_columns(status: "firing")

    get "/api/v1/#{@account.slug}/beeps",
      params: { status: "firing" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    beeps = response.parsed_body["beeps"]
    assert_equal 1, beeps.size
    assert_equal firing_beep.id, beeps.first["id"]
  end

  test "index filters by kind" do
    once_beep = @account.beeps.create!(kind: :once, title: "Once Beep", run_at: @run_at)
    recurring_beep = @account.beeps.create!(kind: :recurring, title: "Recurring Beep", cron: "0 9 * * *")

    get "/api/v1/#{@account.slug}/beeps",
      params: { kind: "recurring" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    beeps = response.parsed_body["beeps"]
    assert_equal 1, beeps.size
    assert_equal recurring_beep.id, beeps.first["id"]
  end

  test "create with :in duration schedules run_at in the future" do
    travel_to Time.utc(2026, 9, 23, 10, 0, 0) do
      post "/api/v1/#{@account.slug}/beeps",
        params: { title: "Tea break", in: "15m" },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json

      assert_response :created
      body = response.parsed_body
      assert_equal "Tea break", body["title"]
      assert_equal "once", body["kind"]
      assert_equal "active", body["status"]
      expected_run_at = Time.utc(2026, 9, 23, 10, 15, 0)
      assert_equal expected_run_at.iso8601, Time.iso8601(body["run_at"]).iso8601
      assert_equal expected_run_at.iso8601, Time.iso8601(body["next_run_at"]).iso8601
    end
  end

  test "create with :in duration rejects invalid format" do
    post "/api/v1/#{@account.slug}/beeps",
      params: { title: "Bad in", in: "not-a-duration" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert_equal "VALIDATION_ERROR", response.parsed_body["code"]
    assert_match(/Invalid in duration/, response.parsed_body["message"])
  end

  test "create with :at time-only schedules run_at respecting user timezone" do
    users(:john).update!(timezone: "Asia/Shanghai", timezone_source: "manual")

    # Current time: 10:00 UTC, user timezone Asia/Shanghai (18:00 local)
    travel_to Time.utc(2026, 9, 23, 10, 0, 0) do
      # Target: 19:30 local today (which is 11:30 UTC today)
      # Request payload timezone cannot override user's configured timezone
      post "/api/v1/#{@account.slug}/beeps",
        params: { title: "Dinner", at: "19:30", timezone: "America/New_York" },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json

      assert_response :created
      body = response.parsed_body
      assert_equal "Asia/Shanghai", body["timezone"]
      assert_equal "once", body["kind"]
      expected_utc = Time.utc(2026, 9, 23, 11, 30, 0)
      assert_equal expected_utc.iso8601, Time.iso8601(body["run_at"]).iso8601

      # Target: 12:00 local (already passed today in Asia/Shanghai 18:00 -> rolls to tomorrow 12:00 local = tomorrow 04:00 UTC)
      post "/api/v1/#{@account.slug}/beeps",
        params: { title: "Lunch tomorrow", at: "12:00" },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json

      assert_response :created
      body = response.parsed_body
      expected_tomorrow_utc = Time.utc(2026, 9, 24, 4, 0, 0)
      assert_equal expected_tomorrow_utc.iso8601, Time.iso8601(body["run_at"]).iso8601
    end
  end

  test "create with :at falls back to request timezone when user timezone is unset" do
    assert_nil users(:john).timezone

    travel_to Time.utc(2026, 9, 23, 10, 0, 0) do
      post "/api/v1/#{@account.slug}/beeps",
        params: { title: "Dinner", at: "19:30", timezone: "Asia/Shanghai" },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json

      assert_response :created
      body = response.parsed_body
      assert_equal "Asia/Shanghai", body["timezone"]
      expected_utc = Time.utc(2026, 9, 23, 11, 30, 0)
      assert_equal expected_utc.iso8601, Time.iso8601(body["run_at"]).iso8601
    end
  end

  test "create with :at full datetime schedules run_at" do
    travel_to Time.utc(2026, 9, 23, 10, 0, 0) do
      post "/api/v1/#{@account.slug}/beeps",
        params: { title: "Conference", at: "2026-10-01 15:00", timezone: "Asia/Shanghai" },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json

      assert_response :created
      body = response.parsed_body
      expected_utc = Time.utc(2026, 10, 1, 7, 0, 0)
      assert_equal expected_utc.iso8601, Time.iso8601(body["run_at"]).iso8601
    end
  end

  test "create rejects conflicting schedule options" do
    # in and at
    post "/api/v1/#{@account.slug}/beeps",
      params: { title: "Conflict", in: "15m", at: "15:00" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert_equal "VALIDATION_ERROR", response.parsed_body["code"]
    assert_match(/Cannot specify multiple schedule options/, response.parsed_body["message"])

    # run_at and in
    post "/api/v1/#{@account.slug}/beeps",
      params: { title: "Conflict", in: "15m", run_at: 1.hour.from_now.iso8601 },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert_match(/Cannot specify multiple schedule options/, response.parsed_body["message"])

    # cron and in
    post "/api/v1/#{@account.slug}/beeps",
      params: { title: "Conflict", in: "15m", cron: "0 9 * * *" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert_match(/Cannot specify :cron together with :in or :at/, response.parsed_body["message"])
  end

  test "update supports :in and :at to reschedule beep" do
    beep = @account.beeps.create!(kind: :once, title: "Meeting", run_at: @run_at, timezone: "UTC")

    travel_to Time.utc(2026, 9, 23, 10, 0, 0) do
      patch "/api/v1/#{@account.slug}/beeps/#{beep.id}",
        params: { in: "45m" },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json

      assert_response :success
      body = response.parsed_body
      expected_run_at = Time.utc(2026, 9, 23, 10, 45, 0)
      assert_equal expected_run_at.iso8601, Time.iso8601(body["run_at"]).iso8601
      assert_equal expected_run_at.to_i, beep.reload.run_at.to_i

      # Reschedule with :at and transient timezone:
      # schedules run_at in Asia/Shanghai local time while leaving stored beep.timezone unchanged ("UTC")
      patch "/api/v1/#{@account.slug}/beeps/#{beep.id}",
        params: { at: "19:30", timezone: "Asia/Shanghai" },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json

      assert_response :success
      body = response.parsed_body
      expected_at_utc = Time.utc(2026, 9, 23, 11, 30, 0)
      assert_equal expected_at_utc.iso8601, Time.iso8601(body["run_at"]).iso8601
      assert_equal expected_at_utc.to_i, beep.reload.run_at.to_i
      assert_equal "UTC", body["timezone"]
      assert_equal "UTC", beep.reload.timezone
    end
  end
end
