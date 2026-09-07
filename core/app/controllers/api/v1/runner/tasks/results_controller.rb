class Api::V1::Runner::Tasks::ResultsController < Api::V1::Runner::BaseController
  before_action :set_run

  def create
    raw_metrics = params[:metrics]
    metrics = if raw_metrics.respond_to?(:to_unsafe_h)
      raw_metrics.to_unsafe_h
    elsif raw_metrics.is_a?(Hash)
      raw_metrics
    else
      {}
    end

    recorded = @run.record_result!(
      status: params[:status],
      title: params[:title],
      message: params[:message],
      metrics: metrics
    )

    unless recorded
      return render_json_error(
        status: :unprocessable_entity,
        message: "Run already has a result",
        code: "VALIDATION_ERROR"
      )
    end

    @current_runner.touch_activity(status: "idle")
    render :create
  end
end
