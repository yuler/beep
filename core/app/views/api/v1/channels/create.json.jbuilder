json.channel do
  json.id @channel.id
  json.name @channel.name
  json.kind @channel.kind
  json.status @channel.status
  json.token @channel.token
  json.masked_token @channel.masked_token
  json.user_id @channel.user_id
  json.last_seen_at @channel.last_seen_at
  json.created_at @channel.created_at
end
