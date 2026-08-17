class BeepPollerJob < ApplicationJob
  def perform
    Beep.poll_due_now
  end
end
