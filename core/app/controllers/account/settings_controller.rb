class Account::SettingsController < ApplicationController
  before_action :ensure_admin

  def show
    @account = Current.account
  end

  def update
    @account = Current.account

    if @account.update(account_params)
      redirect_to account_settings_url(script_name: @account.slug_path),
        notice: "Account settings updated."
    else
      render :show, status: :unprocessable_entity
    end
  end

  private
    def account_params
      params.expect(account: [ :name, :slug, :description ])
    end
end
