class Api::V1::BeeperAppsController < Api::V1::BaseController
  allow_unauthenticated_access only: %i[ index show ]
  skip_account_scope only: %i[ index show ]

  def index
    @beeper_apps = BeeperApp.official.order(:slug)
    render :index
  end

  def show
    @beeper_app = BeeperApp.official.find_by!(slug: params[:id])
    render :show
  end
end
