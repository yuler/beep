class Api::V1::Runner::DisconnectsController < Api::V1::Runner::BaseController
  def destroy
    @current_runner.destroy!
    head :no_content
  end
end
