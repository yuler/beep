class Api::V1::Runner::Jobs::PushesController < Api::V1::Runner::BaseController
  def create
    jobs_payload = Array(params[:jobs])
    @pushed_jobs = []

    ActiveRecord::Base.transaction do
      jobs_payload.each do |job_data|
        attrs = job_attrs_from(job_data)
        next if attrs[:slug].blank?

        job = find_or_build_job(id: job_data[:id], slug: attrs[:slug])
        job.assign_attributes(attrs)
        job.save!
        @pushed_jobs << job
      end
    end

    render template: "api/v1/runner/jobs/pushes/create"
  rescue ActiveRecord::RecordInvalid => e
    render_json_error(
      status: :unprocessable_entity,
      message: e.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end
end
