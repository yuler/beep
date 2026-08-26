json.plugins @plugins do |plugin|
  json.partial! "api/v1/plugins/plugin", plugin: plugin
end
