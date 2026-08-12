class LandingsController < ApplicationController
  allow_unauthenticated_access
  # Account-scoped (/:slug) and global landing both allowed — skip require, do not redirect when account is set.
  skip_before_action :require_account
  before_action :require_account_context_when_authenticated

  def show
    render "home/index" unless authenticated?
  end

  private
    def require_account_context_when_authenticated
      if authenticated? && Current.account.blank?
        redirect_to my_accounts_url
      end
    end
end
