json.beeper_installs @beeper_installs do |install|
  json.partial! "api/v1/beeper_installs/beeper_install", beeper_install: install
end
