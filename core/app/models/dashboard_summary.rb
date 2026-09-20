class DashboardSummary
  attr_reader :account, :range, :now, :since

  def initialize(account, range: "7d", now: Time.current)
    @account = account
    @range = range.in?(%w[ 24h 7d ]) ? range : "7d"
    @now = now
    @since = @range == "24h" ? 24.hours.ago(now) : 7.days.ago(now)
  end

  def stats
    {
      executions: execution_stats,
      beeps: beep_stats,
      beepers: beeper_stats,
      runners: runner_stats
    }
  end

  def chart_points
    buckets = build_buckets

    # Populate beep runs
    account.beep_runs.where(created_at: since..now).pluck(:created_at, :status).each do |created_at, status|
      key = bucket_key_for(created_at)
      next unless (b = buckets[key])
      b[:total] += 1
      if status == "succeeded"
        b[:success] += 1
      elsif status.in?(%w[ failed expired ])
        b[:failure] += 1
      end
    end

    # Populate beeper runs
    account.beeper_runs.where(created_at: since..now).pluck(:created_at, :status, :signal_status).each do |created_at, status, signal_status|
      key = bucket_key_for(created_at)
      next unless (b = buckets[key])
      b[:total] += 1
      if status == "succeeded" && signal_status == "ok"
        b[:success] += 1
      elsif status.in?(%w[ failed expired ]) || signal_status.in?(%w[ alerting error ])
        b[:failure] += 1
      end
    end

    # Populate runner runs
    account.runner_runs.where(created_at: since..now).pluck(:created_at, :status, :result_status).each do |created_at, status, result_status|
      key = bucket_key_for(created_at)
      next unless (b = buckets[key])
      b[:total] += 1
      if status == "succeeded" && (result_status == "ok" || result_status.blank?)
        b[:success] += 1
      elsif status.in?(%w[ failed expired ]) || result_status == "error"
        b[:failure] += 1
      end
    end

    buckets.values
  end

  def activities(limit: 20)
    items = []

    account.beep_runs.includes(:beep).order(created_at: :desc).limit(limit).each do |run|
      next unless run.beep
      items << {
        id: "beep_run_#{run.id}",
        type: "beep",
        title: run.beep.title,
        status: beep_run_status(run),
        summary: "Execution #{run.status}",
        occurred_at: (run.scheduled_for || run.created_at).iso8601,
        target_path: "beeps/#{run.beep_id}"
      }
    end

    account.beeper_runs.includes(:beeper).order(created_at: :desc).limit(limit).each do |run|
      next unless run.beeper
      items << {
        id: "beeper_run_#{run.id}",
        type: "beeper",
        title: run.beeper.title,
        status: beeper_run_status(run),
        summary: "Probe #{run.signal_status || run.status}",
        occurred_at: (run.scheduled_for || run.created_at).iso8601,
        target_path: "beepers/#{run.beeper_id}"
      }
    end

    account.runner_runs.includes(:runner_job, :runner).order(created_at: :desc).limit(limit).each do |run|
      next unless run.runner_job
      runner_name = run.runner&.name || "Runner"
      items << {
        id: "runner_run_#{run.id}",
        type: "runner",
        title: run.runner_job.name,
        status: runner_run_status(run),
        summary: "Executed on #{runner_name} (#{run.result_status || run.status})",
        occurred_at: run.created_at.iso8601,
        target_path: "runners/#{run.runner_id}"
      }
    end

    items.sort_by { |item| item[:occurred_at] }.reverse.first(limit)
  end

  private

  def execution_stats
    total = 0
    success = 0
    failure = 0

    # Beep runs
    beep_counts = account.beep_runs.where(created_at: since..now).group(:status).count
    beep_counts.each do |st, cnt|
      total += cnt
      if st == "succeeded"
        success += cnt
      elsif st.in?(%w[ failed expired ])
        failure += cnt
      end
    end

    # Beeper runs
    account.beeper_runs.where(created_at: since..now).group(:status, :signal_status).count.each do |(st, sig), cnt|
      total += cnt
      if st == "succeeded" && sig == "ok"
        success += cnt
      elsif st.in?(%w[ failed expired ]) || sig.in?(%w[ alerting error ])
        failure += cnt
      end
    end

    # Runner runs
    account.runner_runs.where(created_at: since..now).group(:status, :result_status).count.each do |(st, res), cnt|
      total += cnt
      if st == "succeeded" && (res == "ok" || res.blank?)
        success += cnt
      elsif st.in?(%w[ failed expired ]) || res == "error"
        failure += cnt
      end
    end

    rate = total.positive? ? ((success.to_f / total) * 100).round(2) : 100.0

    {
      total: total,
      success_count: success,
      failure_count: failure,
      success_rate: rate
    }
  end

  def beep_stats
    Beep.stats_for(account, now)
  end

  def beeper_stats
    beepers = account.beepers
    total = beepers.count
    healthy = beepers.where(alert_state: "ok").count
    alerting = beepers.where.not(alert_state: "ok").count

    {
      total: total,
      healthy: healthy,
      alerting: alerting
    }
  end

  def runner_stats
    runners = account.runners
    total = runners.count
    online = runners.where(status: "online").count
    active_jobs = account.runner_jobs.where(status: "active").count

    {
      total: total,
      online: online,
      active_jobs: active_jobs
    }
  end

  def build_buckets
    buckets = {}
    if range == "24h"
      start_hour = now.beginning_of_hour - 23.hours
      24.times do |i|
        t = start_hour + i.hours
        key = t.strftime("%Y-%m-%d %H:00")
        buckets[key] = {
          timestamp: t.iso8601,
          label: t.strftime("%H:00"),
          total: 0,
          success: 0,
          failure: 0
        }
      end
    else
      start_day = now.beginning_of_day - 6.days
      7.times do |i|
        t = start_day + i.days
        key = t.strftime("%Y-%m-%d")
        buckets[key] = {
          timestamp: t.iso8601,
          label: t.strftime("%b %d"),
          total: 0,
          success: 0,
          failure: 0
        }
      end
    end
    buckets
  end

  def bucket_key_for(time)
    if range == "24h"
      time.beginning_of_hour.strftime("%Y-%m-%d %H:00")
    else
      time.beginning_of_day.strftime("%Y-%m-%d")
    end
  end

  def beep_run_status(run)
    case run.status
    when "succeeded" then "success"
    when "failed", "expired" then "failure"
    when "running" then "running"
    else "pending"
    end
  end

  def beeper_run_status(run)
    if run.signal_status == "ok"
      "success"
    elsif run.signal_status == "alerting"
      "warning"
    elsif run.status.in?(%w[ failed expired ]) || run.signal_status == "error"
      "failure"
    elsif run.status == "running"
      "running"
    else
      "pending"
    end
  end

  def runner_run_status(run)
    if run.result_status == "ok"
      "success"
    elsif run.status.in?(%w[ failed expired ]) || run.result_status == "error"
      "failure"
    elsif run.status == "running"
      "running"
    else
      "pending"
    end
  end
end
