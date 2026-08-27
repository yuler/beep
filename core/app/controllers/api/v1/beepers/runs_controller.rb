class Api::V1::Beepers::RunsController < Api::V1::BaseController
  def create
    @beeper = Current.account.beepers.find(params[:beeper_id])
    @run = @beeper.trigger_run!
    render partial: "api/v1/beepers/run", locals: { run: @run }, status: :created
  end
end
