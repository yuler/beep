class Beep::Run < ApplicationRecord
  self.table_name = "beep_runs"

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

  def deliver_to(channel)
    channel.deliver_beep(beep, run: self)
  end
  alias_method :deliver_to_channel, :deliver_to

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
      self.class.where(id: id, status: :pending).update_all(status: "running", updated_at: Time.current) == 1
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
        raise EmailDeliveryError, payload_result.dig("email", "error") if email_failed?(payload_result)
      end

      if channels.include?("cli") || channels.any? { |c| c.to_s.start_with?("cli:") }
        payload_result = deliver_cli(user, channels, payload_result)
        persist_result(payload_result)
      end

      channel_ids = channels.select { |c| c.to_s.match?(/\A[0-9a-zA-Z_-]{10,40}\z/) && !c.to_s.in?(User::NOTIFICATION_CHANNELS) }
      if channel_ids.any?
        user.channels.active.where(id: channel_ids).each do |channel|
          payload_result = deliver_explicit_channel(user, channel, payload_result)
          persist_result(payload_result)
          raise EmailDeliveryError, payload_result.dig("email", "error") if email_failed?(payload_result)
        end
      end

      payload_result
    end

    def persist_result(payload_result)
      update_columns(result: payload_result, updated_at: Time.current)
    end

    def deliver_web_push(user, payload_result)
      return payload_result if payload_result.key?("web_push")

      subscriptions = user.channels.active.where(kind: :web_push).to_a
      if subscriptions.empty?
        payload_result.merge("web_push" => { "reason" => "no_subscriptions" })
      else
        payload_result.merge("web_push" => { "deliveries" => subscriptions.map { |ch| deliver_to(ch) } })
      end
    end

    def deliver_email(user, payload_result)
      return payload_result if email_attempt_complete?(payload_result)

      send_email(user, payload_result)
    end

    def email_attempt_complete?(payload_result)
      payload_result.dig("email", "status").in?(%w[ sent skipped ])
    end

    def email_failed?(payload_result)
      payload_result.dig("email", "status") == "error"
    end

    def send_email(user, payload_result)
      BeepMailer.beep(self, user: user).deliver_now
      payload_result.merge("email" => { "status" => "sent" })
    rescue StandardError => error
      payload_result.merge("email" => { "status" => "error", "error" => error.class.name })
    end

    def deliver_cli(user, channels, payload_result)
      return payload_result if payload_result.key?("cli")

      target_channels = user.channels.active.where(kind: :cli)
      specific_names = channels.select { |c| c.to_s.start_with?("cli:") }.map { |c| c.to_s.delete_prefix("cli:") }
      target_channels = target_channels.where(name: specific_names) if specific_names.any?
      target_channels = target_channels.to_a

      if target_channels.empty?
        payload_result.merge("cli" => { "reason" => "no_channels" })
      else
        payload_result.merge("cli" => { "deliveries" => target_channels.map { |ch| deliver_to(ch) } })
      end
    end

    def deliver_explicit_channel(user, channel, payload_result)
      if channel.kind == "email"
        deliver_email(user, payload_result)
      else
        delivery_record = deliver_to(channel)
        existing = payload_result.dig(channel.kind, "deliveries") || []
        payload_result.merge(channel.kind => { "deliveries" => existing + [ delivery_record ] })
      end
    end
end
