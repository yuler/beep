class Api::V1::Runner::Tasks::LogsController < Api::V1::Runner::BaseController
  before_action :set_run

  MAX_CHUNK_BYTES = 512.kilobytes

  def create
    unless @run.running? || @run.pending?
      return render_json_error(
        status: :unprocessable_entity,
        message: "Run is no longer accepting logs",
        code: "VALIDATION_ERROR"
      )
    end

    chunk = params[:chunk].to_s
    if chunk.blank? && params[:lines].is_a?(Array)
      chunk = Array(params[:lines]).join("\n")
      chunk = "#{chunk}\n" if chunk.present?
    end

    if chunk.bytesize > MAX_CHUNK_BYTES
      chunk = chunk.byteslice(0, MAX_CHUNK_BYTES)
    end

    @run.append_log(chunk)
    @current_runner.touch_activity(status: "online")
    head :no_content
  end
end
