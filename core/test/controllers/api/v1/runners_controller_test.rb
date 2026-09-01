require "test_helper"

class Api::V1::RunnersControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
  end

  test "index returns account runners" do
    runner = @account.runners.create!(name: "Nas-01", tags: ["intranet"])

    get "/api/v1/#{@account.slug}/runners",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    runners = response.parsed_body["runners"]
    assert_equal 1, runners.size
    assert_equal runner.id, runners.first["id"]
    assert_equal "Nas-01", runners.first["name"]
    assert_equal ["intranet"], runners.first["tags"]
  end

  test "show returns runner details" do
    runner = @account.runners.create!(name: "Nas-01", tags: ["intranet"])

    get "/api/v1/#{@account.slug}/runners/#{runner.id}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body["runner"]
    assert_equal runner.id, body["id"]
    assert_equal "Nas-01", body["name"]
  end

  test "create creates a new runner and returns raw token" do
    assert_difference -> { @account.runners.count }, 1 do
      post "/api/v1/#{@account.slug}/runners",
        params: {
          runner: {
            name: "Office-Gateway",
            tags: ["office", "intranet"],
            allow_exec: true
          }
        },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json

      assert_response :created
      body = response.parsed_body["runner"]
      assert_equal "Office-Gateway", body["name"]
      assert_equal ["office", "intranet"], body["tags"]
      assert_equal true, body["allow_exec"]
      assert body["token"].start_with?("beep_rt_")
    end
  end

  test "update updates runner attributes" do
    runner = @account.runners.create!(name: "Old-Name")

    put "/api/v1/#{@account.slug}/runners/#{runner.id}",
      params: {
        runner: {
          name: "New-Name",
          tags: ["production"],
          allow_exec: true
        }
      },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    runner.reload
    assert_equal "New-Name", runner.name
    assert_equal ["production"], runner.tags
    assert_equal true, runner.allow_exec
  end

  test "destroy removes runner" do
    runner = @account.runners.create!(name: "To-Delete")

    assert_difference -> { @account.runners.count }, -1 do
      delete "/api/v1/#{@account.slug}/runners/#{runner.id}",
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json

      assert_response :no_content
    end
  end

  test "regenerate_token issues a new token" do
    runner = @account.runners.create!(name: "Token-Runner")
    old_digest = runner.token_digest

    post "/api/v1/#{@account.slug}/runners/#{runner.id}/regenerate_token",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :success
    body = response.parsed_body["runner"]
    assert body["token"].start_with?("beep_rt_")
    runner.reload
    assert_not_equal old_digest, runner.token_digest
  end
end
