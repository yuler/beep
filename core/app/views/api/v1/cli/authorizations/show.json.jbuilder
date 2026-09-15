json.user_code @auth.user_code
json.client_name @auth.client_name
json.status @auth.status
json.expires_in [ (@auth.expires_at - Time.current).to_i, 0 ].max
