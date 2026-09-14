class Api::V1::Runner::ConnectionsController < Api::V1::Runner::BaseController
  def destroy
    @current_runner.destroy!
    head :no_content
  end
end
