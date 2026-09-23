class Api::V1::BeepPreviewsController < Api::V1::BaseController
  def create
    @preview = Beep::Preview.build(
      account: Current.account,
      user: Current.user,
      params: preview_params
    )

    render :create, status: :ok
  end

  private
    def preview_params
      params.permit(:title, :body, :in, :at, :run_at, :cron, :kind, :timezone, :intent, notification_channels: [], metadata: {})
    end
end
