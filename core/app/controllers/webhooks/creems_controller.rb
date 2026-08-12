class Webhooks::CreemsController < ApplicationController
  disallow_account_scope
  allow_unauthenticated_access
  skip_forgery_protection

  def create
    if body = verify_signature
      save_webhook_event(body)
      dispatch_webhook_event(body)
      head :ok
    else
      head :bad_request
    end
  rescue ArgumentError => e
    Rails.logger.error "Creem webhook processing error: #{e.message}"
    head :internal_server_error
  end

  private
    # refs: https://docs.creem.io/code/webhooks#webhook-signatures
    def verify_signature
      payload = request.body.read
      request.body.rewind

      signature = request.headers["Creem-Signature"]
      secret = ENV["CREEM_WEBHOOK_SECRET"]

      if signature.blank? || secret.blank?
        raise ArgumentError, "Missing Creem-Signature or CREEM_WEBHOOK_SECRET"
      end

      expected = OpenSSL::HMAC.hexdigest("SHA256", secret, payload)
      raise ActiveSupport::MessageVerifier::InvalidSignature, "Creem webhook signature mismatch" unless ActiveSupport::SecurityUtils.secure_compare(expected, signature)

      JSON.parse(payload)
    end

    # save to database
    def save_webhook_event(body)
      event_id = body["id"]

      if webhook = Account::PaymentWebhook.find_by(provider: :creem, event_id: event_id)
        webhook
      else
        event_type = body["eventType"]
        account_id = body.dig("object", "metadata", "account_id")
        account = Account.find_by!(id: account_id)
        Account::PaymentWebhook.create!(
          account: account,
          provider: :creem,
          event_type: event_type,
          event_id: event_id,
          raw: body.to_json
        )
      end
    end

    def dispatch_webhook_event(body)
      event_type = body["eventType"]
      checkout_id = body.dig("object", "id")

      case event_type
      when "checkout.completed"
        charge = Account::Payable.find_charge(provider: :creem, checkout_id: checkout_id)
        charge&.update!(status: "succeeded")
      else
        Rails.logger.warn "Unhandled event type: #{event_type}"
      end
    end
end
