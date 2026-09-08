class Api::V1::Runner::PingsController < Api::V1::Runner::BaseController
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

    render :create
  end
end
