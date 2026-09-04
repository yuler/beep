class Api::V1::Runner::TasksController < Api::V1::Runner::BaseController
  def poll
    @current_runner.touch_activity!(
      status: "idle",
      version: params[:version],
      os: params[:os],
      arch: params[:arch],
      hostname: params[:hostname],
      ip_address: request.remote_ip,
      allow_exec: params[:allow_exec]
    )

    candidate = RunnerRun.joins(:runner_job)
                         .where(runner_jobs: { account_id: @current_runner.account_id, runner_id: @current_runner.id, status: "firing" })
                         .where(status: "pending")
                         .where(scheduled_for: ..Time.current)
                         .order("runner_runs.scheduled_for ASC")
                         .first

    if candidate&.claim_for!(@current_runner)
      @current_runner.update_columns(status: "online")
      @run = candidate
      render :poll
      return
    end

    head :no_content
  end

  def logs
    @run = find_run
    unless @run.running? || @run.pending?
      return render_json_error(
        status: :unprocessable_entity,
        message: "Run is no longer accepting logs",
        code: "VALIDATION_ERROR"
      )
    end

    chunk = params[:chunk].to_s
    if chunk.blank? && params[:lines].is_a?(Array)
      chunk = Array(params[:lines]).join("\n")
      chunk = "#{chunk}\n" if chunk.present?
    end

    @run.append_log!(chunk)
    @current_runner.touch_activity!(status: "online")
    render :logs
  end

  def result
    @run = find_run

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

    @current_runner.touch_activity!(status: "idle")
    render :result
  end

  private

  def find_run
    RunnerRun.joins(:runner_job)
             .where(runner_jobs: { account_id: @current_runner.account_id })
             .where(runner_id: @current_runner.id)
             .find(params[:id])
  end
end
