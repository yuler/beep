class Api::V1::Runner::JobsController < Api::V1::Runner::BaseController
  def index
    @jobs = @current_runner.jobs.order(:name)
    render :index
  end

  def create
    attrs = job_attrs_from(params)

    @job = find_or_build_job(id: params[:id], slug: attrs[:slug])
    @job.assign_attributes(attrs)

    if @job.save
      render :show, status: :created
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @job.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  def destroy
    slug = params[:id].to_s.strip.downcase
    @job = @current_runner.jobs.find_by(slug: slug) || @current_runner.jobs.find_by(id: params[:id])

    if @job
      @job.destroy
      head :no_content
    else
      render_json_error(
        status: :not_found,
        message: "Job not found",
        code: "NOT_FOUND"
      )
    end
  end
end
