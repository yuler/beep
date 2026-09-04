json.status "ok"
json.runner_id @current_runner.id
json.runner_name @current_runner.name
json.server_time Time.current.utc.iso8601
