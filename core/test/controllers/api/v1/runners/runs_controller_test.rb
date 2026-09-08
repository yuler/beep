require "test_helper"

class Api::V1::Runners::RunsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = identities(:john)
    @account = accounts(:john_account)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @runner = @account.runners.create!(name: "Office")
    @job = @runner.jobs.create!(
      name: "Intranet HTTP",
      slug: "intranet-http",
      cron: "*/5 * * * *",
      timezone: "UTC",
      timeout_seconds: 20
    )
  end

  test "destroy removes a terminal run" do
    run = @job.trigger_run!
    run.update!(status: :succeeded)

    delete "/api/v1/#{@account.slug}/runners/#{@runner.id}/jobs/#{@job.id}/runs/#{run.id}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :no_content
    assert_not Runner::Run.exists?(run.id)
  end

  test "destroy rejects a pending run" do
    run = @job.trigger_run!

    delete "/api/v1/#{@account.slug}/runners/#{@runner.id}/jobs/#{@job.id}/runs/#{run.id}",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :unprocessable_entity
    assert Runner::Run.exists?(run.id)
  end

  test "clear removes terminal runs and keeps pending ones" do
    succeeded = @job.trigger_run!
    succeeded.update!(status: :succeeded)
    failed = @job.trigger_run!
    failed.update!(status: :failed)
    pending = @job.trigger_run!

    delete "/api/v1/#{@account.slug}/runners/#{@runner.id}/jobs/#{@job.id}/runs/history",
      headers: { "Authorization" => "Bearer #{@token}" },
      as: :json

    assert_response :no_content
    assert_not Runner::Run.exists?(succeeded.id)
    assert_not Runner::Run.exists?(failed.id)
    assert Runner::Run.exists?(pending.id)
  end
end
