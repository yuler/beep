class DeliverBeepRunJob < ApplicationJob
  retry_on BeepRun::EmailDeliveryError, wait: 15.seconds, attempts: 5 do |job, _error|
    beep_run = job.arguments.first
    if beep_run.running?
      beep_run.update!(status: :failed)
      beep_run.beep.finish_firing(last_run_at: beep_run.scheduled_for)
    end
  end

  def perform(beep_run)
    beep_run.deliver_now
  end
end
