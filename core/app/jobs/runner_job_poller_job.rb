class RunnerJobPollerJob < ApplicationJob
  def perform
    RunnerJob.poll_due_now
  end
end
