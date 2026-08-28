json.beeper_apps @beeper_apps do |beeper_app|
  json.partial! "api/v1/beeper_apps/beeper_app", beeper_app: beeper_app
end
