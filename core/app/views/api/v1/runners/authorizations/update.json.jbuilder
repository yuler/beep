json.status "approved"
json.runner do
  json.id @runner.id
  json.name @runner.name
  json.tags @runner.tags
  json.status @runner.status
  json.masked_token @runner.masked_token
end
