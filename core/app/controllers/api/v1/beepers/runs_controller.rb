class Api::V1::Beepers::RunsController < Api::V1::BaseController
  before_action :set_beeper

  def index
    @runs = set_page_and_extract_portion_from @beeper.runs, ordered_by: { scheduled_for: :desc, id: :desc }
    render :index
  end

  def create
    @run = @beeper.trigger_run!
    render partial: "api/v1/beepers/run", locals: { run: @run }, status: :created
  end

  private
    def set_beeper
      @beeper = Current.account.beepers.find(params[:beeper_id])
    end
end
