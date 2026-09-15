class Api::V1::Channels::Cli::InboxesController < Api::V1::Channels::Cli::BaseController
  def show
    @current_channel.deliveries.expire_stale!
    @current_channel.deliveries.reclaim_stale!
    due = @current_channel.deliveries.due_for_cli.order(:created_at).limit(10).to_a
    claimed_ids = []
    due.each do |delivery|
      claimed_ids << delivery.id if delivery.claim!
    end
    @deliveries = @current_channel.deliveries.where(id: claimed_ids).order(:created_at)
    render :show
  end
end
