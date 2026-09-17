json.beeps @beeps do |beep|
  json.partial! "api/v1/beeps/beep",
    beep: beep,
    run_stats: @run_stats.fetch(beep.id, { total: 0, succeeded: 0 }),
    runs: @recent_runs.fetch(beep.id, [])
end
json.partial! "api/v1/shared/pagination", page: @page
