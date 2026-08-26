class Api::V1::PluginsController < Api::V1::BaseController
  allow_unauthenticated_access only: %i[ index show ]
  skip_account_scope only: %i[ index show ]

  def index
    @plugins = Plugin.official.order(:slug)
    render :index
  end

  def show
    @plugin = Plugin.official.find_by!(slug: params[:id])
    render :show
  end
end
