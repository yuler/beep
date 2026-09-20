json.stats do
  json.executions do
    json.total @stats[:executions][:total]
    json.success_count @stats[:executions][:success_count]
    json.failure_count @stats[:executions][:failure_count]
    json.success_rate @stats[:executions][:success_rate]
  end

  json.beeps do
    json.active @stats[:beeps][:active]
    json.due_today @stats[:beeps][:due_today]
    json.firing @stats[:beeps][:firing]
    json.recurring @stats[:beeps][:recurring]
    json.completed @stats[:beeps][:completed]
    json.all @stats[:beeps][:all]
  end

  json.beepers do
    json.total @stats[:beepers][:total]
    json.healthy @stats[:beepers][:healthy]
    json.alerting @stats[:beepers][:alerting]
  end

  json.runners do
    json.total @stats[:runners][:total]
    json.online @stats[:runners][:online]
    json.active_jobs @stats[:runners][:active_jobs]
  end
end

json.chart do
  json.range @range
  json.points @chart_points do |point|
    json.timestamp point[:timestamp]
    json.label point[:label]
    json.total point[:total]
    json.success point[:success]
    json.failure point[:failure]
  end
end

json.activities @activities do |activity|
  json.id activity[:id]
  json.type activity[:type]
  json.title activity[:title]
  json.status activity[:status]
  json.summary activity[:summary]
  json.occurred_at activity[:occurred_at]
  json.target_path activity[:target_path]
end
