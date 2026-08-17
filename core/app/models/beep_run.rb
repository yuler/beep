class BeepRun < ApplicationRecord
  belongs_to :beep

  enum :status, %w[ pending running succeeded failed skipped expired ].index_by(&:itself)

  def deliver_later
    DeliverBeepRunJob.perform_later(self)
  end

  def deliver_now
    return if succeeded? || failed? || skipped? || expired?

    update!(status: :running)
    payload_result = deliver_web_push
    update!(status: :succeeded, result: payload_result)
    beep.finish_firing(last_run_at: scheduled_for)
  end

  private
    def deliver_web_push
      subscriptions = Push::Subscription.where(account_id: beep.account_id).to_a
      if subscriptions.empty?
        { "web_push" => { "reason" => "no_subscriptions" } }
      else
        {
          "web_push" => {
            "deliveries" => subscriptions.map { |subscription| deliver_to(subscription) }
          }
        }
      end
    end

    def deliver_to(subscription)
      subscription.deliver_beep(beep)
      { "subscription_id" => subscription.id, "status" => "sent" }
    rescue WebPush::ExpiredSubscription, WebPush::InvalidSubscription
      subscription.destroy!
      { "subscription_id" => subscription.id, "status" => "expired" }
    rescue StandardError => error
      { "subscription_id" => subscription.id, "status" => "error", "error" => error.class.name }
    end
end
