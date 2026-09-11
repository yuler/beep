class BeepRun < ApplicationRecord
  class EmailDeliveryError < StandardError; end

  belongs_to :beep

  enum :status, %w[ pending running succeeded failed skipped expired ].index_by(&:itself)

  def deliver_later
    DeliverBeepRunJob.perform_later(self)
  end

  def deliver_now
    deliver_notifications_now if claim_delivery?
  end

  def fail_now
    update!(status: :failed)
    beep.finish_firing(last_run_at: scheduled_for)
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

    def deliver_for(user, payload_result)
      channels = Array(beep.notification_channels)
      return payload_result if channels.empty?

      if channels.include?("web_push")
        payload_result = deliver_web_push(user, payload_result)
        persist_result(payload_result)
      end

      if channels.include?("email")
        payload_result = deliver_email(user, payload_result)
        persist_result(payload_result)
        if payload_result.dig("email", "status") == "error"
          raise EmailDeliveryError, payload_result.dig("email", "error")
        end
      end

      if channels.include?("cli") || channels.any? { |c| c.to_s.start_with?("cli:") }
        payload_result = deliver_cli(user, channels, payload_result)
        persist_result(payload_result)
      end

      channel_ids = channels.select { |c| c.to_s.match?(/\A[0-9a-zA-Z_-]{10,40}\z/) && !c.to_s.in?(User::NOTIFICATION_CHANNELS) }
      if channel_ids.any?
        user_channels = user.channels.active.where(id: channel_ids).to_a
        user_channels.each do |channel|
          case channel.kind
          when "email"
            payload_result = deliver_email(user, payload_result)
            persist_result(payload_result)
            if payload_result.dig("email", "status") == "error"
              raise EmailDeliveryError, payload_result.dig("email", "error")
            end
          when "web_push"
            delivery_record = deliver_to(channel)
            existing = payload_result.dig("web_push", "deliveries") || []
            payload_result = payload_result.merge("web_push" => { "deliveries" => existing + [ delivery_record ] })
            persist_result(payload_result)
          when "cli"
            delivery_record = deliver_to_channel(channel)
            existing = payload_result.dig("cli", "deliveries") || []
            payload_result = payload_result.merge("cli" => { "deliveries" => existing + [ delivery_record ] })
            persist_result(payload_result)
          end
        end
      end

      payload_result
    end

    def persist_result(payload_result)
      update_columns(result: payload_result, updated_at: Time.current)
    end

    def deliver_web_push(user, payload_result)
      if payload_result.key?("web_push")
        payload_result
      else
        payload_result.merge(web_push_payload(user))
      end
    end

    def web_push_payload(user)
      subscriptions = user.channels.active.where(kind: :web_push).to_a
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
      { "subscription_id" => subscription.id, "channel_id" => subscription.id, "status" => "sent" }
    rescue WebPush::ExpiredSubscription, WebPush::InvalidSubscription
      subscription.destroy!
      { "subscription_id" => subscription.id, "channel_id" => subscription.id, "status" => "expired" }
    rescue StandardError => error
      { "subscription_id" => subscription.id, "channel_id" => subscription.id, "status" => "error", "error" => error.class.name }
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

    def deliver_cli(user, channels, payload_result)
      if payload_result.key?("cli")
        payload_result
      else
        payload_result.merge(cli_payload(user, channels))
      end
    end

    def cli_payload(user, channels)
      target_channels = user.channels.active.where(kind: :cli)
      specific_names = channels.select { |c| c.to_s.start_with?("cli:") }.map { |c| c.to_s.delete_prefix("cli:") }
      if specific_names.any?
        target_channels = target_channels.where(name: specific_names)
      end

      if target_channels.empty?
        { "cli" => { "reason" => "no_channels" } }
      else
        {
          "cli" => {
            "deliveries" => target_channels.map { |channel| deliver_to_channel(channel) }
          }
        }
      end
    end

    def deliver_to_channel(channel)
      expires = (scheduled_for || Time.current) + ChannelDelivery::DEFAULT_TTL
      delivery = channel.deliveries.create!(
        beep_run: self,
        payload: {
          id: id,
          event: "beep.fired",
          source: beep.source_slug,
          source_type: beep.source_type,
          source_id: beep.source_id,
          intent: beep.intent,
          beep_id: beep.id,
          title: beep.title,
          body: beep.body_text,
          scheduled_for: (scheduled_for || Time.current).iso8601,
          expires_at: expires.iso8601,
          metadata: beep.metadata || {}
        },
        expires_at: expires
      )
      { "channel_id" => channel.id, "channel_name" => channel.name, "delivery_id" => delivery.id, "status" => "queued" }
    rescue StandardError => error
      { "channel_id" => channel.id, "channel_name" => channel.name, "status" => "error", "error" => error.class.name }
    end
end
