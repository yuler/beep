class HomeController < ApplicationController
  disallow_account_scope
  allow_unauthenticated_access

  def index
  end
end
