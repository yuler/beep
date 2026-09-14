class Api::V1::Channels::Cli::ConnectionsController < Api::V1::Channels::Cli::BaseController
  def destroy
    @current_channel.destroy!
    head :no_content
  end
end
