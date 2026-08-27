json.beepers @beepers do |beeper|
  json.partial! "api/v1/beepers/beeper", beeper: beeper
end
