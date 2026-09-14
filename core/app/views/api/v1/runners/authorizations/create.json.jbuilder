json.device_code @auth.device_code
json.user_code @auth.user_code
if @account_slug.present?
  json.verification_uri "#{@web_origin}/#{@account_slug}/device/runner"
  json.verification_uri_complete "#{@web_origin}/#{@account_slug}/device/runner?code=#{@auth.user_code}"
else
  json.verification_uri "#{@web_origin}/device/runner"
  json.verification_uri_complete "#{@web_origin}/device/runner?code=#{@auth.user_code}"
end
json.expires_in (@auth.expires_at - Time.current).to_i
json.interval Runner::Authorization::DEFAULT_INTERVAL
