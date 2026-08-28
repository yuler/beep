class Api::V1::BeepsController < Api::V1::BaseController
  def index
    @beeps = Current.account.beeps.includes(:runs, beeper: :beeper_app).order(created_at: :desc)
    render :index
  end

  def show
    @beep = Current.account.beeps.includes(:runs, beeper: :beeper_app).find(params[:id])
    render :show
  end

  def create
    kind = params[:kind].presence || (params[:cron].present? ? "recurring" : "once")
    @beep = Current.account.beeps.new(beep_params.merge(kind: kind, timezone: beep_timezone))

    if @beep.save
      render :create, status: :created
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @beep.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  def update
    @beep = Current.account.beeps.find(params[:id])

    if @beep.update(beep_params)
      render :show
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @beep.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  def destroy
    Current.account.beeps.find(params[:id]).destroy!
    head :no_content
  end

  private
    def beep_params
      params.permit(:title, :body, :run_at, :cron, :kind, notification_channels: [])
    end

    def beep_timezone
      IanaTimezone.resolve(Current.user.timezone, params[:timezone])
    end
end
