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

  private
    def upsert_push_subscription
      subscription = Current.user.push_subscriptions.create_with(subscription_attributes)
        .create_or_find_by!(endpoint: subscription_params[:endpoint])
      subscription.update!(subscription_attributes)
      subscription
    end

    def subscription_attributes
      subscription_params.merge(user_agent: request.user_agent.to_s.truncate(4096))
    end

    def subscription_params
      params.permit(:endpoint, :p256dh_key, :auth_key)
    end
end
