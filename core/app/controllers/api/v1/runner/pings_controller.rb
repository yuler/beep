class Api::V1::Runner::PingsController < Api::V1::Runner::BaseController
  def create
    @current_runner.touch_activity!(
      version: params[:version],
      os: params[:os],
      arch: params[:arch],
      hostname: params[:hostname],
      ip_address: request.remote_ip,
      allow_exec: params[:allow_exec]
    )

    render :create
  end
end
