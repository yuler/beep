json.access_token @access_token.token
json.token_type "bearer"
json.user do
  json.id @identity.id
  json.email @identity.email
  json.name @identity.full_name
end
if (last_slug = @identity.last_account_slug.presence)
  json.account_slug last_slug
elsif (personal = @identity.personal_account)
  json.account_slug personal.slug
end
