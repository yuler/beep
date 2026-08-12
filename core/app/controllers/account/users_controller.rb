class Account::UsersController < ApplicationController
  def index
    @users = Current.account.users
  end

  def new
    @user = Current.account.users.new
  end
end
