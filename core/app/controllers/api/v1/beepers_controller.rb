class Api::V1::BeepersController < Api::V1::BaseController
  def index
    @beepers = Current.account.beepers.includes(:beeper_app, :runs).order(created_at: :desc)
    render :index
  end

  def show
    @beeper = Current.account.beepers.includes(:beeper_app, :runs).find(params[:id])
    render :show
  end

  def create
    beeper_app = find_beeper_app
    unless beeper_app
      return render_json_error(
        status: :unprocessable_entity,
        message: "Beeper app not found or not official",
        code: "VALIDATION_ERROR"
      )
    end

    cron = params[:cron].presence || beeper_app.default_cron
    @beeper = Current.account.beepers.new(
      beeper_params.merge(
        beeper_app: beeper_app,
        cron: cron,
        timezone: install_timezone
      )
    )

    if @beeper.save
      render :create, status: :created
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @beeper.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  def update
    @beeper = Current.account.beepers.find(params[:id])

    if @beeper.update(update_params)
      render :show
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @beeper.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  def destroy
    Current.account.beepers.find(params[:id]).destroy!
    head :no_content
  end

  private

  def find_beeper_app
    if params[:beeper_app_id].present?
      BeeperApp.official.find_by(id: params[:beeper_app_id])
    elsif params[:beeper_app_slug].present?
      BeeperApp.official.find_by(slug: params[:beeper_app_slug])
    elsif params[:beeper_id].present?
      # backward compat fallback if needed
      BeeperApp.official.find_by(id: params[:beeper_id])
    elsif params[:beeper_slug].present?
      # backward compat fallback if needed
      BeeperApp.official.find_by(slug: params[:beeper_slug])
    end
  end

  def beeper_params
    attrs = params.permit(:title, :cron, notification_channels: [])
    attrs[:config] = params[:config].to_unsafe_h if params[:config].respond_to?(:to_unsafe_h)
    attrs
  end

  def update_params
    attrs = params.permit(:title, :cron, notification_channels: [])
    attrs[:config] = params[:config].to_unsafe_h if params[:config].respond_to?(:to_unsafe_h)
    attrs
  end

  def install_timezone
    IanaTimezone.resolve(Current.user.timezone, params[:timezone])
  end
end
