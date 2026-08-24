json.access_token do
  json.partial! "api/v1/my/access_tokens/access_token", access_token: @access_token
  json.token @access_token.token
end
