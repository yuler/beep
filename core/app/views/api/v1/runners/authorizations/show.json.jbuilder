json.user_code @auth.user_code
json.runner_name @auth.runner_name
json.tags @auth.tags
json.metadata @auth.metadata
json.status @auth.status
json.expires_in [ (@auth.expires_at - Time.current).to_i, 0 ].max
