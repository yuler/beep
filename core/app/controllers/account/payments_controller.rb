class Account::PaymentsController < ApplicationController
  before_action :ensure_admin
  before_action :set_charge, only: :show

  def show
    redirect_to home_index_path(anchor: "pricing") if @charge.blank?
  end

  def create
    success_url = account_payment_url(provider: plan_param.provider)
    @charge = Account::Payable.create_charge(plan_key: plan_param.key, success_url: success_url)
    redirect_to @charge.raw_json.fetch("checkout_url"), allow_other_host: true
  rescue Account::Payable::Creem::CreemError => e
    Rails.logger.error("[PaymentsController] Payment creation failed: #{e.message}")
    flash[:alert] = "Failed to create payment: #{e.message}"
    redirect_back_or_to root_path
  end

  private
    def plan_param
      @plan_param ||= Account::Payable::Plan.find(params[:plan_key]) || Account::Payable::Plan.paid
    end

    def set_charge
      if params[:provider].present? && params[:checkout_id].present?
        @charge = Current.account.charges.find_by(
          provider: params[:provider],
          checkout_id: params[:checkout_id],
        )
      end
    end
end
