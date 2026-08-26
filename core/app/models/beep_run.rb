class BeepRun < ApplicationRecord
  class EmailDeliveryError < StandardError; end

  CHECK_RESULT_MAX_BYTES = 8.kilobytes

  belongs_to :beep

  enum :status, %w[ pending running succeeded failed skipped expired ].index_by(&:itself)
  enum :check_status, %w[ ok alerting error ].index_by(&:itself)

  def deliver_later
    if beep.plugin?
      RunCheckJob.perform_later(self)
    else
      DeliverBeepRunJob.perform_later(self)
    end
  end

  def execute_check_now
    return unless claim_delivery?

    if beep.expired_plugin_run?(scheduled_for)
      update!(status: :expired)
      beep.finish_firing(last_run_at: scheduled_for)
      return
    end

    check_result = beep.plugin.run_check(config: beep.effective_plugin_config)
    sanitized_result = sanitize_check_result(check_result.to_h)

    decision = BeepRun::AlertEvaluator.evaluate(
      beep: beep,
      check_result: check_result
    )

    update!(
      check_status: check_result.status.to_s,
      check_result: sanitized_result
    )

    beep.update!(
      alert_state: decision.next_alert_state,
      consecutive_failures: decision.next_consecutive_failures
    )

    if decision.should_notify
      # Trigger delivery for alert / recovery notification
      deliver_notifications_now
    else
      update!(status: :succeeded)
      beep.finish_firing(last_run_at: scheduled_for)
    end
  rescue StandardError => e
    update!(
      check_status: "error",
      check_result: { "status" => "error", "message" => e.message },
      status: :failed
    )
    beep.finish_firing(last_run_at: scheduled_for)
  end

  def deliver_now
    return unless claim_delivery?

    deliver_notifications_now
  end

  private

    def deliver_notifications_now
      payload_result = stringify_result
      beep.recipient_users.each do |user|
        payload_result = deliver_for(user, payload_result)
      end

      update!(status: :succeeded, result: payload_result)
      beep.finish_firing(last_run_at: scheduled_for)
    end

    def claim_delivery?
      claimed = self.class.where(id: id, status: :pending).update_all(status: "running", updated_at: Time.current) == 1
      claimed || running?
    end

    def stringify_result
      (result || {}).deep_stringify_keys
    end

    def sanitize_check_result(hash)
      json_str = hash.to_json
      if json_str.bytesize > CHECK_RESULT_MAX_BYTES
        {
          "status" => hash["status"],
          "title" => hash["title"],
          "message" => hash["message"]&.truncate(500),
          "metrics" => hash["metrics"],
          "truncated" => true
        }.compact
      else
        hash
      end
    end

    def persist_result(payload_result)
      update_columns(result: payload_result, updated_at: Time.current)
    end

    def deliver_for(user, payload_result)
      if user.notification_channel?("web_push")
        payload_result = deliver_web_push(user, payload_result)
        persist_result(payload_result)
      end

      if user.notification_channel?("email")
        payload_result = deliver_email(user, payload_result)
        persist_result(payload_result)
        if payload_result.dig("email", "status") == "error"
          raise EmailDeliveryError, payload_result.dig("email", "error")
        end
      end

      payload_result
    end

    def deliver_web_push(user, payload_result)
      if payload_result.key?("web_push")
        payload_result
      else
        payload_result.merge(web_push_payload(user))
      end
    end

    def web_push_payload(user)
      subscriptions = user.push_subscriptions.to_a
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
      subscription.deliver_beep(beep, run: self)
      { "subscription_id" => subscription.id, "status" => "sent" }
    rescue WebPush::ExpiredSubscription, WebPush::InvalidSubscription
      subscription.destroy!
      { "subscription_id" => subscription.id, "status" => "expired" }
    rescue StandardError => error
      { "subscription_id" => subscription.id, "status" => "error", "error" => error.class.name }
    end

    def deliver_email(user, payload_result)
      if email_attempt_complete?(payload_result)
        payload_result
      else
        send_reminder_email(user, payload_result)
      end
    end

    def email_attempt_complete?(payload_result)
      payload_result.dig("email", "status").in?(%w[ sent skipped ])
    end

    def send_reminder_email(user, payload_result)
      BeepMailer.reminder(self, user: user).deliver_now
      payload_result.merge("email" => { "status" => "sent" })
    rescue StandardError => error
      payload_result.merge("email" => { "status" => "error", "error" => error.class.name })
    end
end
