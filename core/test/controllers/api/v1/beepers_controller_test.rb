require "test_helper"

class Api::V1::BeepersControllerTest < ActionDispatch::IntegrationTest
  include ActiveJob::TestHelper

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
          body: "Check customer checkout gateway",
          cron: "*/5 * * * *",
          config: { "target_url" => "https://example.com" }
        },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :created
    body = response.parsed_body
    assert_equal "example.com", body["title"]
    assert_equal "Check customer checkout gateway", body["body"]
    assert_equal "site-uptime", body["beeper_app"]["slug"]
    assert_nil body["kind"]
  end

  test "update modifies beeper properties including notification_channels" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Old Title",
      body: "Old Body",
      cron: "*/5 * * * *",
      timezone: "UTC",
      notification_channels: [ "email" ],
      config: { "target_url" => "https://example.com" }
    )

    patch "/api/v1/#{@account.slug}/beepers/#{beeper.id}",
      params: {
        title: "New Title",
        body: "Updated Body",
        cron: "*/10 * * * *",
        notification_channels: [ "email", "web_push" ],
        config: { "target_url" => "https://new.example.com", "expected_status" => 201 }
      },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    res_body = response.parsed_body
    assert_equal "New Title", res_body["title"]
    assert_equal "Updated Body", res_body["body"]
    assert_equal "*/10 * * * *", res_body["cron"]
    assert_equal [ "email", "web_push" ], res_body["notification_channels"]
    assert_equal "https://new.example.com", res_body["config"]["target_url"]

    beeper.reload
    assert_equal "New Title", beeper.title
    assert_equal "Updated Body", beeper.body
    assert_equal "*/10 * * * *", beeper.cron
    assert_equal [ "email", "web_push" ], beeper.notification_channels
    assert_equal "https://new.example.com", beeper.config["target_url"]
  end

  test "update clears body when set to null" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "With Body",
      body: "Old Body",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    patch "/api/v1/#{@account.slug}/beepers/#{beeper.id}",
      params: { body: nil },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_nil response.parsed_body["body"]
    assert_nil beeper.reload.body
  end

  test "update clears body when set to empty string" do
    beeper = Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "With Body",
      body: "Old Body",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    patch "/api/v1/#{@account.slug}/beepers/#{beeper.id}",
      params: { body: "" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_nil response.parsed_body["body"]
    assert_nil beeper.reload.body
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

  test "create allows installing custom beeper app with a receiver belonging to the account" do
    custom_app = BeeperApp.create!(
      account: @account,
      slug: "echo",
      version: "1.0.0",
      manifest: {
        "manifest_version" => 1,
        "slug" => "echo",
        "name" => "Echo Checker",
        "version" => "1.0.0",
        "author" => "Me",
        "schedule" => { "default_cron" => "*/10 * * * *" }
      }
    )

    assert_difference -> { @account.beepers.count }, 1 do
      post "/api/v1/#{@account.slug}/beepers",
        headers: { "Authorization" => "Bearer #{@token}" },
        params: {
          beeper_app_id: custom_app.id,
          title: "My Custom Probe",
          config: { "status" => "ok" }
        },
        as: :json
    end

    assert_response :created
    beeper = @account.beepers.order(:created_at).last
    assert_equal custom_app.id, beeper.beeper_app_id
    assert_equal "My Custom Probe", beeper.title
  end

  test "create rejects custom beeper app without a receiver implementation" do
    custom_app = BeeperApp.create!(
      account: @account,
      slug: "unimplemented-checker",
      version: "1.0.0",
      manifest: {
        "manifest_version" => 1,
        "slug" => "unimplemented-checker",
        "name" => "Unimplemented Checker",
        "version" => "1.0.0",
        "author" => "Me",
        "schedule" => { "default_cron" => "*/10 * * * *" }
      }
    )

    assert_no_difference -> { @account.beepers.count } do
      post "/api/v1/#{@account.slug}/beepers",
        headers: { "Authorization" => "Bearer #{@token}" },
        params: {
          beeper_app_id: custom_app.id,
          title: "My Custom Probe",
          config: {}
        },
        as: :json
    end

    assert_response :unprocessable_entity
    assert_match(/no receiver implementation/i, response.parsed_body["message"])
  end

  test "create rejects another account custom beeper app" do
    other_account = accounts(:yuler_account)
    other_app = BeeperApp.create!(
      account: other_account,
      slug: "other-checker",
      version: "1.0.0",
      manifest: {
        "manifest_version" => 1,
        "slug" => "other-checker",
        "name" => "Other Checker",
        "version" => "1.0.0",
        "author" => "Other",
        "schedule" => { "default_cron" => "*/10 * * * *" }
      }
    )

    assert_no_difference -> { @account.beepers.count } do
      post "/api/v1/#{@account.slug}/beepers",
        headers: { "Authorization" => "Bearer #{@token}" },
        params: {
          beeper_app_id: other_app.id,
          title: "Cross Tenant Probe",
          config: {}
        },
        as: :json
    end

    assert_response :unprocessable_entity
    assert_equal "Beeper app not found", response.parsed_body["message"]
  end

  test "show returns not found for another account beeper" do
    other_account = accounts(:yuler_account)
    beeper = Beeper.create!(
      account: other_account,
      beeper_app: @beeper_app,
      title: "Other Account Beeper",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    get "/api/v1/#{@account.slug}/beepers/#{beeper.id}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
  end

  test "update returns not found for another account beeper" do
    other_account = accounts(:yuler_account)
    beeper = Beeper.create!(
      account: other_account,
      beeper_app: @beeper_app,
      title: "Other Account Beeper",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    patch "/api/v1/#{@account.slug}/beepers/#{beeper.id}",
      params: { title: "Hacked" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :not_found
    assert_equal "Other Account Beeper", beeper.reload.title
  end

  test "destroy returns not found for another account beeper" do
    other_account = accounts(:yuler_account)
    beeper = Beeper.create!(
      account: other_account,
      beeper_app: @beeper_app,
      title: "Other Account Beeper",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    assert_no_difference -> { Beeper.count } do
      delete "/api/v1/#{@account.slug}/beepers/#{beeper.id}",
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :not_found
  end

  test "runs create returns not found for another account beeper" do
    other_account = accounts(:yuler_account)
    beeper = Beeper.create!(
      account: other_account,
      beeper_app: @beeper_app,
      title: "Other Account Beeper",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    assert_no_enqueued_jobs only: RunBeeperJob do
      post "/api/v1/#{@account.slug}/beepers/#{beeper.id}/runs",
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :not_found
  end

  test "index excludes beepers from other accounts" do
    Beeper.create!(
      account: @account,
      beeper_app: @beeper_app,
      title: "Mine",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )
    Beeper.create!(
      account: accounts(:yuler_account),
      beeper_app: @beeper_app,
      title: "Theirs",
      cron: "*/5 * * * *",
      timezone: "UTC",
      config: { "target_url" => "https://example.com" }
    )

    get "/api/v1/#{@account.slug}/beepers",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    titles = response.parsed_body["beepers"].map { |row| row["title"] }
    assert_equal [ "Mine" ], titles
  end
end
