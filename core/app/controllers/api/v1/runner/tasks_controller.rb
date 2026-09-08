class Api::V1::Runner::TasksController < Api::V1::Runner::BaseController
  def create
    has_running = @current_runner.runs.where(status: "running").exists?
    @current_runner.touch_activity(
      status: has_running ? "online" : "idle",
      version: params[:version],
      os: params[:os],
      arch: params[:arch],
      hostname: params[:hostname],
      ip_address: request.remote_ip
    )

    candidate = Runner::Run.joins(:runner_job)
                         .where(runner_jobs: { account_id: @current_runner.account_id, runner_id: @current_runner.id })
                         .where(status: "pending")
                         .where(scheduled_for: ..Time.current)
                         .order("runner_runs.scheduled_for ASC")
                         .first

    if candidate&.claim_for(@current_runner)
      @current_runner.update_columns(status: "online")
      @run = candidate
      @api_base_url = runner_callback_base_url
      render :create
    else
      head :no_content
    end
  end

  private
    # Prefer configured Core host over request Host to avoid forged callback URLs.
    def runner_callback_base_url
      opts = Rails.application.config.action_mailer.default_url_options || {}
      host = opts[:host]
      if host.blank?
        request.base_url
      else
        protocol = opts[:protocol].presence || (Rails.env.local? ? "http" : "https")
        port = opts[:port]
        if port.present? && ![ 80, 443 ].include?(port.to_i)
          "#{protocol}://#{host}:#{port}"
        else
          "#{protocol}://#{host}"
        end
      end
    end
end
