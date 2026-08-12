class Api::V1::Admin::JobsController < Api::V1::Admin::BaseController
  def show
    render_json json: {
      adapter: ActiveJob::Base.queue_adapter_name.to_s,
      available: solid_queue_available?,
      counts: solid_queue_counts,
      recent: solid_queue_recent_jobs
    }
  end

  private
    def solid_queue_available?
      defined?(SolidQueue::Job) && SolidQueue::Job.table_exists?
    rescue ActiveRecord::StatementInvalid, ActiveRecord::ConnectionNotEstablished
      false
    end

    def solid_queue_counts
      return nil unless solid_queue_available?

      {
        total: SolidQueue::Job.count,
        finished: SolidQueue::Job.where.not(finished_at: nil).count,
        pending: SolidQueue::Job.where(finished_at: nil).count,
        failed: SolidQueue::Job.failed.count,
        ready: SolidQueue::ReadyExecution.count,
        scheduled: SolidQueue::ScheduledExecution.count
      }
    end

    def solid_queue_recent_jobs
      return [] unless solid_queue_available?

      SolidQueue::Job.order(created_at: :desc).limit(20).map { |job|
        {
          id: job.id,
          class_name: job.class_name,
          queue_name: job.queue_name,
          finished_at: job.finished_at&.iso8601,
          created_at: job.created_at.iso8601,
          failed: job.failed_execution.present?
        }
      }
    end
end
