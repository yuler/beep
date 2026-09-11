json.channels @channels do |channel|
  json.id channel.id
  json.name channel.name
  json.kind channel.kind
  json.status channel.status
  json.is_online channel.online?
  json.masked_token channel.masked_token
  json.user do
    json.id channel.user_id
    json.name channel.user&.name
  end
  json.last_seen_at channel.last_seen_at
  json.created_at channel.created_at
end
