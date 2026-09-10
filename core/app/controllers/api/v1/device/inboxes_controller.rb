class Api::V1::Device::InboxesController < Api::V1::Device::BaseController
  def show
    @deliveries = @current_device.deliveries.due_for_device.order(:created_at).limit(10)
    @deliveries.each(&:claim!)
    render :show
  end
end
