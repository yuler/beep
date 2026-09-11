json.user_code @auth.user_code
json.channel_name @auth.channel_name
json.status @auth.status
json.expires_in [ (@auth.expires_at - Time.current).to_i, 0 ].max
