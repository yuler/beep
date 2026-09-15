json.access_token @access_token.token
json.token_type "bearer"
json.user do
  json.id @identity.id
  json.email @identity.email
  json.name @identity.full_name
end
