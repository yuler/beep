class Api::V1::PushSubscriptions::TestsController < Api::V1::BaseController
  before_action :set_push_subscription

  def create
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
    def set_push_subscription
      @push_subscription = Current.user.push_subscriptions.find(params[:push_subscription_id])
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
        message: "This browser is no longer subscribed",
        code: "PUSH_SUBSCRIPTION_EXPIRED"
      )
    end
end
