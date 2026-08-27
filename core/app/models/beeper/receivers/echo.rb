class Beeper::Receivers::Echo < Beeper::Receivers::Base
  def call
    status = config["status"] || "ok"
    Beeper::Signal.new(
      status: status,
      title: config["title"] || "Echo Signal",
      message: config["message"] || "Echo signal received",
      metrics: config["metrics"] || { "echo" => true }
    )
  end
end
