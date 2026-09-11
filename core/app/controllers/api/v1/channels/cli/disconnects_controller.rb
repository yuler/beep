class Api::V1::Channels::Cli::DisconnectsController < Api::V1::Channels::Cli::BaseController
  def destroy
    @current_channel.destroy!
    head :no_content
  end
end
