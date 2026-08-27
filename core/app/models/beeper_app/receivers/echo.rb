class BeeperApp::Receivers::Echo < BeeperApp::Receivers::Base
  def call
    status = config["status"] || "ok"
    BeeperApp::Signal.new(
      status: status,
      title: config["title"] || "Echo Signal",
      message: config["message"] || "Echo signal received",
      metrics: config["metrics"] || { "echo" => true }
    )
  end
end
