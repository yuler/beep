json.status "approved"
json.channel do
  json.id @channel.id
  json.name @channel.name
  json.kind @channel.kind
  json.status @channel.status
  json.masked_token @channel.masked_token
end
