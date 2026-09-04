class Api::V1::RunnersController < Api::V1::BaseController
  before_action :set_runner, only: %i[ show update destroy regenerate_token ]

  def index
    Runner.mark_stale_offline
    @runners = Current.account.runners.order(created_at: :desc)
    render :index
  end

  def show
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
    if @runner.update(update_params)
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

  def regenerate_token
    @runner.regenerate_token
    render :create
  end

  private

    def set_runner
      @runner = Current.account.runners.find(params[:id])
    end

    def runner_params
      params.require(:runner).permit(:name, tags: [])
    end

    def update_params
      params.require(:runner).permit(:name, tags: [])
    end
end
