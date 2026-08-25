class Api::V1::Beeps::PausesController < Api::V1::BaseController
  before_action :set_beep

  def create
    @beep.pause!
    render "api/v1/beeps/show"
  end

  def destroy
    @beep.resume!
    render "api/v1/beeps/show"
  end

  private
    def set_beep
      @beep = Current.account.beeps.find(params[:beep_id])
    end
end
