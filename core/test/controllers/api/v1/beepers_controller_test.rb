require "test_helper"

class Api::V1::BeepersControllerTest < ActionDispatch::IntegrationTest
  setup do
    BeeperApp.seed_official
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @beeper_app = BeeperApp.find_by!(slug: "site-uptime")
  end

  test "index returns account beepers" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "My Uptime",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    get "/api/v1/#{@account.slug}/beepers",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    beepers = response.parsed_body["beepers"]
    assert_equal 1, beepers.size
    assert_equal beeper.id, beepers.first["id"]
    assert_equal "My Uptime", beepers.first["title"]
    assert_equal "site-uptime", beepers.first["beeper_app"]["slug"]
  end

  test "show returns beeper details" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "My Uptime",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    get "/api/v1/#{@account.slug}/beepers/#{beeper.id}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal beeper.id, body["id"]
    assert_equal "My Uptime", body["title"]
    assert_equal "site-uptime", body["beeper_app"]["slug"]
  end

  test "create creates a new beeper via beeper_app_slug" do
    assert_difference -> { @account.beepers.count }, 1 do
      post "/api/v1/#{@account.slug}/beepers",
        params: {
          beeper_app_slug: "site-uptime",
          title: "example.com",
          cron: "*/5 * * * *",
          config: { "target_url" => "https://example.com" }
        },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :created
    body = response.parsed_body
    assert_equal "example.com", body["title"]
    assert_equal "site-uptime", body["beeper_app"]["slug"]
    assert_nil body["kind"]
  end

  test "update modifies beeper properties" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Old Title",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    patch "/api/v1/#{@account.slug}/beepers/#{beeper.id}",
      params: { title: "New Title" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "New Title", response.parsed_body["title"]
    assert_equal "New Title", beeper.reload.title
  end

  test "destroy deletes the beeper" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "To Delete",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    assert_difference -> { @account.beepers.count }, -1 do
      delete "/api/v1/#{@account.slug}/beepers/#{beeper.id}",
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :no_content
  end

  test "pause and resume updates beeper status via pauses controller" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "To Pause",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    post "/api/v1/#{@account.slug}/beepers/#{beeper.id}/pause",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "paused", response.parsed_body["status"]
    assert beeper.reload.paused?

    delete "/api/v1/#{@account.slug}/beepers/#{beeper.id}/pause",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "active", response.parsed_body["status"]
    assert beeper.reload.active?
  end

  test "runs create triggers a new run for an existing beeper" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Trigger Test",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    assert_enqueued_with(job: RunBeeperJob) do
      assert_difference -> { beeper.runs.count }, 1 do
        post "/api/v1/#{@account.slug}/beepers/#{beeper.id}/runs",
          headers: { "Authorization" => "Bearer #{@token}" },
          as: :json
      end
    end

    assert_response :created
    body = response.parsed_body
    assert_equal "pending", body["status"]
    assert beeper.reload.firing?
  end
end
