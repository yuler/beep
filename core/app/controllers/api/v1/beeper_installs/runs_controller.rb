class Api::V1::BeeperInstalls::RunsController < Api::V1::BaseController
  def create
    @beeper_install = Current.account.beeper_installs.find(params[:beeper_install_id])
    @run = @beeper_install.trigger_run!
    render partial: "api/v1/beeper_installs/run", locals: { run: @run }, status: :created
  end
end
