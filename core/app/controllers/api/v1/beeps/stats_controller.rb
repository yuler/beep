class Api::V1::Beeps::StatsController < Api::V1::BaseController
  def show
    @stats = Beep.stats_for(Current.account)
    render :show
  end
end
