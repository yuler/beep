class Api::V1::BeeperApps::Heartbeat::PingsController < ActionController::API
  CACHE_STORE = Class.new {
    def self.increment(...)
      Rails.cache.increment(...)
    end
  }

  rate_limit to: 60, within: 1.minute, by: -> { params[:token].presence || request.remote_ip }, with: -> { head :too_many_requests }, store: CACHE_STORE

  def create
    token = params[:token].to_s.strip
    if token.blank?
      head :bad_request
    elsif (beeper = Beeper.find_by_ping_token(token, beeper_app_slug: params[:id]))
      beeper.record_ping
      head :ok
    else
      head :not_found
    end
  end
end
