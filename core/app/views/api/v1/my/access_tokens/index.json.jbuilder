json.access_tokens @access_tokens do |access_token|
  json.partial! "api/v1/my/access_tokens/access_token", access_token: access_token
end
