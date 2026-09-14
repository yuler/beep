require "test_helper"

class Api::V1::Runner::ConnectionsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @account = accounts(:john_account)
    @runner = @account.runners.create!(name: "One")
  end

  test "destroys current runner when authorized" do
    assert_difference -> { Runner.count }, -1 do
      delete "/api/v1/runner/connection",
        headers: { "X-Runner-Token" => @runner.token }

      assert_response :no_content
    end

    assert_nil Runner.find_by(id: @runner.id)
  end

  test "returns 401 when token is missing or invalid" do
    delete "/api/v1/runner/connection"
    assert_response :unauthorized

    delete "/api/v1/runner/connection",
      headers: { "X-Runner-Token" => "invalid_token" }
    assert_response :unauthorized
  end
end
