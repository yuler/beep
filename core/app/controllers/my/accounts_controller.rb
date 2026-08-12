class My::AccountsController < ApplicationController
  disallow_account_scope

  def new
    @account = Current.identity.accounts.new
  end

  def create
    @account = Account.create_with_owner(
      account: account_params,
      owner: {
        name: Current.identity.full_name,
        identity: Current.identity
      }
    )

    if @account.persisted?
      redirect_to my_accounts_path
    else
      render :new, status: :unprocessable_entity
    end
  end

  def index
    @accounts = Current.identity.accounts
    @admin_account_ids = Current.identity.users.admin.pluck(:account_id)
    @last_account_slug = Current.identity.last_account_slug
  end

  private
    def account_params
      params.expect(account: [ :name, :description ])
    end
end
