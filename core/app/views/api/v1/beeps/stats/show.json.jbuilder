json.stats do
  json.active @stats[:active]
  json.due_today @stats[:due_today]
  json.firing @stats[:firing]
end
