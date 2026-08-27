class Api::V1::BeeperInstalls::PingsController < ActionController::API
  def create
    token = params[:token].to_s.strip
    if token.blank?
      head :bad_request
    elsif (install = BeeperInstall.find_by(ping_token: token))
      install.record_ping
      head :ok
    else
      head :not_found
    end
  end
end
