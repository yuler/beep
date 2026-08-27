class Api::V1::BeeperInstallsController < Api::V1::BaseController
  def index
    @beeper_installs = Current.account.beeper_installs.includes(:beeper).order(created_at: :desc)
    render :index
  end

  def show
    @beeper_install = Current.account.beeper_installs.includes(:beeper).find(params[:id])
    render :show
  end

  def create
    beeper = find_beeper
    unless beeper
      return render_json_error(
        status: :unprocessable_entity,
        message: "Beeper not found or not official",
        code: "VALIDATION_ERROR"
      )
    end

    cron = params[:cron].presence || beeper.default_cron
    @beeper_install = Current.account.beeper_installs.new(
      install_params.merge(
        beeper: beeper,
        cron: cron,
        timezone: install_timezone
      )
    )

    if @beeper_install.save
      render :create, status: :created
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @beeper_install.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  def update
    @beeper_install = Current.account.beeper_installs.find(params[:id])

    if @beeper_install.update(update_params)
      render :show
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @beeper_install.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  def destroy
    Current.account.beeper_installs.find(params[:id]).destroy!
    head :no_content
  end

  private

  def find_beeper
    if params[:beeper_id].present?
      Beeper.official.find_by(id: params[:beeper_id])
    elsif params[:beeper_slug].present?
      Beeper.official.find_by(slug: params[:beeper_slug])
    end
  end

  def install_params
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
