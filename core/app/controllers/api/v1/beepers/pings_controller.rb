class Api::V1::Beepers::PingsController < ActionController::API
  def create
    token = params[:token].to_s.strip
    if token.blank?
      head :bad_request
    elsif (beeper = Beeper.find_by_ping_token(token))
      beeper.record_ping
      head :ok
    else
      head :not_found
    end
  end
end
