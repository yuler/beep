class Api::V1::Channels::TestsController < Api::V1::BaseController
  before_action :set_channel

  def create
    @channel.deliver_test!
    head :no_content
  end

  private
    def set_channel
      @channel = Current.account.channels.find(params[:channel_id])
    end
end
