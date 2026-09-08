class RunnerJobPollerJob < ApplicationJob
  def perform
    Runner::Job.poll_due_now
  end
end
