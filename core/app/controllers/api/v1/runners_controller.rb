class Api::V1::RunnersController < Api::V1::BaseController
  before_action :set_runner, only: %i[ show update destroy ]

  def index
    Runner.mark_stale_offline
    @runners = Current.account.runners
                              .left_joins(:jobs)
                              .select("runners.*, COUNT(runner_jobs.id) AS jobs_count")
                              .group("runners.id")
                              .order(created_at: :desc)
    render :index
  end

  def show
    Runner.mark_stale_offline
    render :show
  end

  def create
    @runner = Current.account.runners.new(runner_params)

    if @runner.save
      render :create, status: :created
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @runner.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  def update
    if @runner.update(runner_params)
      render :show
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @runner.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  def destroy
    @runner.destroy!
    head :no_content
  end

  private
    def set_runner
      @runner = Current.account.runners.find(params[:id])
    end

    def runner_params
      params.require(:runner).permit(:name, tags: [])
    end
end
