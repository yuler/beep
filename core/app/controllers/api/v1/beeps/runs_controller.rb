class Api::V1::Beeps::RunsController < Api::V1::BaseController
  before_action :set_beep

  def index
    @runs = set_page_and_extract_portion_from @beep.runs, ordered_by: { scheduled_for: :desc, id: :desc }
    render :index
  end

  def create
    @run = @beep.trigger_run!
    render :create, status: :created
  end

  private
    def set_beep
      @beep = Current.account.beeps.find(params[:beep_id])
    end
end
