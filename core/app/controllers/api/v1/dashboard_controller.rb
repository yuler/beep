class Api::V1::DashboardController < Api::V1::BaseController
  def show
    @range = params[:range].presence_in(%w[ 24h 7d ]) || "7d"
    @summary = DashboardSummary.new(Current.account, range: @range)
    @stats = @summary.stats
    @chart_points = @summary.chart_points
    @activities = @summary.activities
    render :show
  end
end
