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

    scope = BeeperRun.joins(:beeper)
                     .where(beepers: { account_id: @current_runner.account_id, status: "firing" })
                     .where(status: "pending")
                     .where("beeper_runs.scheduled_for <= ?", Time.current)

    candidate_runs = scope.where(beepers: { runner_id: @current_runner.id })
    if @current_runner.tags.present?
      candidate_runs = candidate_runs.or(scope.where(beepers: { runner_tag: @current_runner.tags }))
    end

    @run = candidate_runs.order("beeper_runs.scheduled_for ASC").first

    if @run
      claimed = BeeperRun.where(id: @run.id, status: "pending").update_all(
        status: "running",
        runner_id: @current_runner.id,
        claimed_at: Time.current,
        updated_at: Time.current
      ) == 1

      if claimed
        @current_runner.update_columns(status: "online")
        @run.reload
        render :poll
        return
      end
    end

    head :no_content
  end

  def result
    @run = BeeperRun.joins(:beeper)
                    .where(beepers: { account_id: @current_runner.account_id })
                    .find(params[:id])

    raw_status = params[:status].to_s.downcase
    status = raw_status.in?(%w[ ok alerting error ]) ? raw_status.to_sym : :error
    title = params[:title].to_s.presence
    message = params[:message].to_s.presence
    raw_metrics = params[:metrics]
    metrics = raw_metrics.respond_to?(:to_unsafe_h) ? raw_metrics.to_unsafe_h : (raw_metrics.is_a?(Hash) ? raw_metrics : {})

    signal = BeeperApp::Signal.new(
      status: status,
      title: title,
      message: message,
      metrics: metrics
    )

    @run.record_signal_result!(signal, runner: @current_runner)
    @current_runner.touch_activity!(status: "idle")

    render :result
  end
end
