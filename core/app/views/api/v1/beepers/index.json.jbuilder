json.beepers @beepers do |beeper|
  json.partial! "api/v1/beepers/beeper",
    beeper: beeper,
    run_stats: @run_stats.fetch(beeper.id, { total: 0, succeeded: 0 }),
    runs: @recent_runs.fetch(beeper.id, [])
end
