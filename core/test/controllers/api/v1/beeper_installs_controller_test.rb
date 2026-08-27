require "test_helper"

class Api::V1::BeeperInstallsControllerTest < ActionDispatch::IntegrationTest
  setup do
    Beeper.seed_official
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @beeper = Beeper.find_by!(slug: "site-uptime")
  end

  test "index returns account installs" do
    install = BeeperInstall.create!(
      account: @account,
      beeper: @beeper,
      title: "My Uptime",
      cron: "*/5 * * * *",
      timezone: "UTC"
    )

    get "/api/v1/#{@account.slug}/beeper_installs",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    installs = response.parsed_body["beeper_installs"]
    assert_equal 1, installs.size
    assert_equal install.id, installs.first["id"]
    assert_equal "My Uptime", installs.first["title"]
    assert_equal "site-uptime", installs.first["beeper"]["slug"]
  end

  test "show returns install details" do
    install = BeeperInstall.create!(
      account: @account,
      beeper: @beeper,
      title: "My Uptime",
      cron: "*/5 * * * *",
      timezone: "UTC"
    )

    get "/api/v1/#{@account.slug}/beeper_installs/#{install.id}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body
    assert_equal install.id, body["id"]
    assert_equal "My Uptime", body["title"]
    assert_equal "site-uptime", body["beeper"]["slug"]
  end

  test "create creates a new beeper install via beeper_slug" do
    assert_difference -> { @account.beeper_installs.count }, 1 do
      post "/api/v1/#{@account.slug}/beeper_installs",
        params: {
          beeper_slug: "site-uptime",
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
    assert_equal "site-uptime", body["beeper"]["slug"]
    assert_nil body["kind"]
  end

  test "update modifies install properties" do
    install = BeeperInstall.create!(
      account: @account,
      beeper: @beeper,
      title: "Old Title",
      cron: "*/5 * * * *",
      timezone: "UTC"
    )

    patch "/api/v1/#{@account.slug}/beeper_installs/#{install.id}",
      params: { title: "New Title" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "New Title", response.parsed_body["title"]
    assert_equal "New Title", install.reload.title
  end

  test "destroy deletes the install" do
    install = BeeperInstall.create!(
      account: @account,
      beeper: @beeper,
      title: "To Delete",
      cron: "*/5 * * * *",
      timezone: "UTC"
    )

    assert_difference -> { @account.beeper_installs.count }, -1 do
      delete "/api/v1/#{@account.slug}/beeper_installs/#{install.id}",
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    end

    assert_response :no_content
  end

  test "pause and resume updates install status via pauses controller" do
    install = BeeperInstall.create!(
      account: @account,
      beeper: @beeper,
      title: "To Pause",
      cron: "*/5 * * * *",
      timezone: "UTC"
    )

    post "/api/v1/#{@account.slug}/beeper_installs/#{install.id}/pause",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "paused", response.parsed_body["status"]
    assert install.reload.paused?

    delete "/api/v1/#{@account.slug}/beeper_installs/#{install.id}/pause",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    assert_equal "active", response.parsed_body["status"]
    assert install.reload.active?
  end

  test "runs create triggers a new run for an existing beeper install" do
    install = BeeperInstall.create!(
      account: @account,
      beeper: @beeper,
      title: "Trigger Test",
      cron: "*/5 * * * *",
      timezone: "UTC"
    )

    assert_enqueued_with(job: RunBeeperJob, queue: "checks") do
      assert_difference -> { install.runs.count }, 1 do
        post "/api/v1/#{@account.slug}/beeper_installs/#{install.id}/runs",
          headers: { "Authorization" => "Bearer #{@token}" },
          as: :json
      end
    end

    assert_response :created
    body = response.parsed_body
    assert_equal "pending", body["status"]
    assert install.reload.firing?
  end

end
