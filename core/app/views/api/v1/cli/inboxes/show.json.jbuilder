json.deliveries @deliveries do |delivery|
  json.id delivery.id
  json.beep_run_id delivery.beep_run_id
  json.status delivery.status
  json.payload delivery.payload
  json.expires_at delivery.expires_at
  json.created_at delivery.created_at
end
