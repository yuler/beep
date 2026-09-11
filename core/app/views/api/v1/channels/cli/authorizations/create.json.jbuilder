json.device_code @auth.device_code
json.user_code @auth.user_code
json.verification_uri "#{@web_origin}/device"
json.verification_uri_complete "#{@web_origin}/device?code=#{@auth.user_code}"
json.expires_in (@auth.expires_at - Time.current).to_i
json.interval ChannelAuthorization::DEFAULT_INTERVAL
