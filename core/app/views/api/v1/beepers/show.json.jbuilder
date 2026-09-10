json.partial! "api/v1/beepers/beeper", beeper: @beeper, runs: @beeper.runs.order(scheduled_for: :desc)
