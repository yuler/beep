class Api::V1::PingsController < ActionController::API
  def create
    token = params[:token].to_s.strip
    if token.blank?
      head :bad_request
      return
    end

    beep = Beep.find_by(ping_token: token)
    if beep.nil?
      head :not_found
      return
    end

    beep.update_columns(last_ping_at: Time.current, updated_at: Time.current)
    head :ok
  end
end
