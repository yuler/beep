class DeliverBeepRunJob < ApplicationJob
  retry_on BeepRun::EmailDeliveryError, wait: 15.seconds, attempts: 5 do |job, _error|
    beep_run = job.arguments.first
    beep_run.fail_now if beep_run.running?
  end

  def perform(beep_run)
    beep_run.deliver_now
  end
end
