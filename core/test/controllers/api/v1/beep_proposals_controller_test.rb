require "test_helper"

class Api::V1::BeepProposalsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @run_at = 1.hour.from_now.change(usec: 0)
    @previous_deepseek_key = ENV["DEEPSEEK_API_KEY"]
    ENV["DEEPSEEK_API_KEY"] = "test-key"
  end

  teardown do
    if @previous_deepseek_key
      ENV["DEEPSEEK_API_KEY"] = @previous_deepseek_key
    else
      ENV.delete("DEEPSEEK_API_KEY")
    end
  end

  test "create returns a proposal without writing a beep" do
    proposal = Beep::Proposal::Result.new(
      intent: "create",
      title: "Call mom",
      body: nil,
      run_at: @run_at,
      timezone: "UTC",
      errors: {},
      message: nil
    )

    original_create = Beep::Proposal.method(:create)
    Beep::Proposal.define_singleton_method(:create) { |*| proposal }

    begin
      assert_no_difference -> { @account.beeps.count } do
        post "/api/v1/#{@account.slug}/beep_proposals",
          params: { prompt: "Tomorrow 9am call mom" },
          headers: { "Authorization" => "Bearer #{@token}" },
          as: :json
      end
    ensure
      Beep::Proposal.define_singleton_method(:create, original_create)
    end

    assert_response :created
    body = response.parsed_body
    assert_equal "create", body["intent"]
    assert_equal "Call mom", body["title"]
    assert_nil body["body"]
    assert_equal @run_at.iso8601, Time.iso8601(body["run_at"]).iso8601
    assert_equal "UTC", body["timezone"]
    assert_equal({}, body["errors"])
    assert body["confirmable"]
    assert_nil body["message"]
  end

  test "create passes the resolved timezone into the proposal" do
    users(:john).update!(timezone: "Europe/London", timezone_source: "manual")
    captured = {}
    run_at = @run_at

    original_create = Beep::Proposal.method(:create)
    Beep::Proposal.define_singleton_method(:create) do |prompt, **kwargs|
      captured[:prompt] = prompt
      captured[:timezone] = kwargs[:timezone]
      Beep::Proposal::Result.new(
        intent: "create",
        title: "Call mom",
        body: nil,
        run_at: run_at,
        timezone: kwargs[:timezone],
        errors: {},
        message: nil
      )
    end

    begin
      post "/api/v1/#{@account.slug}/beep_proposals",
        params: { prompt: "Tomorrow 9am call mom", timezone: "Asia/Tokyo" },
        headers: { "Authorization" => "Bearer #{@token}" },
        as: :json
    ensure
      Beep::Proposal.define_singleton_method(:create, original_create)
    end

    assert_response :created
    assert_equal "Europe/London", captured[:timezone]
    assert_equal "Europe/London", response.parsed_body["timezone"]
  end

  test "create rejects a blank prompt" do
    post "/api/v1/#{@account.slug}/beep_proposals",
      params: { prompt: "  " },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert_equal "VALIDATION_ERROR", response.parsed_body["code"]
  end

  test "create is unavailable without an API key" do
    ENV.delete("DEEPSEEK_API_KEY")

    post "/api/v1/#{@account.slug}/beep_proposals",
      params: { prompt: "Tomorrow 9am call mom" },
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :service_unavailable
    assert_equal "LLM_UNAVAILABLE", response.parsed_body["code"]
  end

  test "create requires authentication" do
    post "/api/v1/#{@account.slug}/beep_proposals",
      params: { prompt: "Tomorrow 9am call mom" },
      as: :json

    assert_response :unauthorized
  end
end
