class BeepRun < ApplicationRecord
  class EmailDeliveryError < StandardError; end

  belongs_to :beep

  enum :status, %w[ pending running succeeded failed skipped expired ].index_by(&:itself)

  def deliver_later
    DeliverBeepRunJob.perform_later(self)
  end

  def deliver_now
    return unless claim_delivery?

    payload_result = stringify_result
    if beep.channel?("web_push")
      payload_result = deliver_web_push(payload_result)
      persist_result(payload_result)
    end

    if beep.channel?("email")
      payload_result = deliver_email(payload_result)
      persist_result(payload_result)
      if payload_result.dig("email", "status") == "error"
        raise EmailDeliveryError, payload_result.dig("email", "error")
      end
    end

    update!(status: :succeeded, result: payload_result)
    beep.finish_firing(last_run_at: scheduled_for)
  end

  private
    def claim_delivery?
      claimed = self.class.where(id: id, status: :pending).update_all(status: "running", updated_at: Time.current) == 1
      claimed || running?
    end

    def stringify_result
      (result || {}).deep_stringify_keys
    end

    def persist_result(payload_result)
      update_columns(result: payload_result, updated_at: Time.current)
    end

    def deliver_web_push(payload_result)
      if payload_result.key?("web_push")
        payload_result
      else
        payload_result.merge(web_push_payload)
      end
    end

    def web_push_payload
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

    def deliver_email(payload_result)
      if email_attempt_complete?(payload_result)
        payload_result
      elsif !beep.account.personal?
        payload_result.merge("email" => { "status" => "skipped", "reason" => "team_account" })
      elsif !beep.account.email_channel_enabled?
        payload_result.merge("email" => { "status" => "skipped", "reason" => "disabled" })
      else
        send_reminder_email(payload_result)
      end
    end

    def email_attempt_complete?(payload_result)
      payload_result.dig("email", "status").in?(%w[ sent skipped ])
    end

    def send_reminder_email(payload_result)
      BeepMailer.reminder(self).deliver_now
      payload_result.merge("email" => { "status" => "sent" })
    rescue StandardError => error
      payload_result.merge("email" => { "status" => "error", "error" => error.class.name })
    end
end
