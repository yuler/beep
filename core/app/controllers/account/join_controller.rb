class Account::JoinController < ApplicationController
  allow_unauthenticated_access only: :show
  allow_unauthorized_access only: :create
  rate_limit to: 10, within: 3.minutes, only: :create, with: -> { head :too_many_requests }

  before_action :set_join_code
  before_action :ensure_join_code_is_valid

  def show
    if Current.identity.blank?
      session[:return_to_after_authenticating] = account_join_url(
        @join_code.code,
        script_name: @join_code.account.slug_path
      )
    end
  end

  def create
    @join_code.redeem_if { |account| Current.identity.join(account) }
    user = User.active.find_by!(account: @join_code.account, identity: Current.identity)
    user.verify

    redirect_to landing_url(script_name: @join_code.account.slug_path)
  end

  private
    def set_join_code
      @join_code = Account::JoinCode.find_by(code: params.expect(:code), account: Current.account)
    end

    def ensure_join_code_is_valid
      if @join_code.nil?
        head :not_found
      elsif !@join_code.active?
        render :inactive, status: :gone
      end
    end
end
