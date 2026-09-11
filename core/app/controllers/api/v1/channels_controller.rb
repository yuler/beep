class Api::V1::ChannelsController < Api::V1::BaseController
  before_action :set_channel, only: %i[ destroy test ]

  def index
    @channels = Current.account.channels.order(created_at: :desc)
    @channels = @channels.where(kind: params[:kind]) if params[:kind].present?
    render :index
  end

  def create
    @channel = Current.account.channels.new(channel_params)
    @channel.user = Current.user

    if @channel.save
      render :create, status: :created
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @channel.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR"
      )
    end
  end

  def destroy
    @channel.destroy!
    head :no_content
  end

  def test
    @channel.deliver_test!
    head :no_content
  end

  private
    def set_channel
      @channel = Current.account.channels.find(params[:id])
    end

    def channel_params
      params.require(:channel).permit(:name, :kind, :config)
    end
end
