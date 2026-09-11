class Api::V1::Channels::Cli::InboxesController < Api::V1::Channels::Cli::BaseController
  def show
    @deliveries = @current_channel.deliveries.due_for_cli.order(:created_at).limit(10)
    @deliveries.each(&:claim!)
    render :show
  end
end
