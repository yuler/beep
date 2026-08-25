class Api::V1::Beeps::RunsController < Api::V1::BaseController
  def create
    @beep = Current.account.beeps.find(params[:beep_id])
    @run = @beep.trigger_run!
    render :create, status: :created
  end
end
