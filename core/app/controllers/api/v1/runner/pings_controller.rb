class Api::V1::Runner::PingsController < Api::V1::Runner::BaseController
  def create
    @current_runner.touch_activity(
      version: params[:version],
      os: params[:os],
      arch: params[:arch],
      hostname: params[:hostname],
      ip_address: request.remote_ip
    )

    render :create
  end
end
