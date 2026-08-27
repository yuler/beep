class Api::V1::BeepersController < Api::V1::BaseController
  allow_unauthenticated_access only: %i[ index show ]
  skip_account_scope only: %i[ index show ]

  def index
    @beepers = Beeper.official.order(:slug)
    render :index
  end

  def show
    @beeper = Beeper.official.find_by!(slug: params[:id])
    render :show
  end
end
