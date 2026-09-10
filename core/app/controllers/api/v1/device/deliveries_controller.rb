class Api::V1::Device::DeliveriesController < Api::V1::Device::BaseController
  def ack
    delivery = @current_device.deliveries.find(params[:id])
    status = params[:status].to_s

    if status == "succeeded"
      delivery.succeed!
      head :no_content
    elsif status == "failed"
      delivery.fail!(params[:error])
      head :no_content
    else
      render_json_error(
        status: :unprocessable_entity,
        message: "Invalid status, must be succeeded or failed",
        code: "VALIDATION_ERROR"
      )
    end
  end
end
