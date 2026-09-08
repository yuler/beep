class Api::V1::Runner::Jobs::PushesController < Api::V1::Runner::BaseController
  def create
    jobs_payload = Array(params[:jobs])
    @pushed_jobs = []

    ActiveRecord::Base.transaction do
      jobs_payload.each do |job_data|
        attrs = job_attrs_from(job_data)
        if attrs[:slug].blank?
          return render_json_error(
            status: :unprocessable_entity,
            message: "Job slug cannot be blank",
            code: "VALIDATION_ERROR"
          )
        end

        job = find_or_build_job(id: job_data[:id], slug: attrs[:slug])
        job.assign_attributes(attrs)
        job.save!
        @pushed_jobs << job
      end
    end
  rescue ActiveRecord::RecordInvalid => e
    render_json_error(
      status: :unprocessable_entity,
      message: e.record.errors.full_messages.to_sentence,
      code: "VALIDATION_ERROR"
    )
  end
end
