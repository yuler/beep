class DeliverBeepRunJob < ApplicationJob
  def perform(beep_run)
    beep_run.deliver_now
  end
end
