class Api::V1::PushSubscriptionsController < Api::V1::BaseController
  def create
    @push_subscription = upsert_push_subscription

    render :create, status: :created
  rescue ActiveRecord::RecordInvalid => error
    render_json_error(
      status: :unprocessable_entity,
      message: error.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end

  def destroy
    Current.user.push_subscriptions.find(params[:id]).destroy!
    head :no_content
  end

  def test
    @push_subscription = Current.user.push_subscriptions.find(params[:id])

    if vapid_configured?
      deliver_test_push
    else
      render_json_error(
        status: :service_unavailable,
        message: "Web Push is not configured",
        code: "WEB_PUSH_UNAVAILABLE"
      )
    end
  end

  private
    def upsert_push_subscription
      Push::Subscription.upsert_for!(Current.user, subscription_attributes)
    end

    def subscription_attributes
      subscription_params.merge(user_agent: request.user_agent.to_s.truncate(4096))
    end

    def subscription_params
      params.permit(:endpoint, :p256dh_key, :auth_key)
    end

    def vapid_configured?
      Rails.application.config.x.vapid.public_key.present? &&
        Rails.application.config.x.vapid.private_key.present?
    end

    def deliver_test_push
      @push_subscription.deliver_test!
      head :no_content
    rescue WebPush::ExpiredSubscription, WebPush::InvalidSubscription
      @push_subscription.destroy!
      render_json_error(
        status: :gone,
        message: "This device is no longer subscribed",
        code: "PUSH_SUBSCRIPTION_EXPIRED"
      )
    end
end
