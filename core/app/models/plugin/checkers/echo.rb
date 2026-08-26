class Plugin::Checkers::Echo < Plugin::Checkers::Base
  def call
    status = config["status"] || "ok"
    Plugin::CheckResult.new(
      status: status,
      title: config["title"] || "Echo Check",
      message: config["message"] || "Echo check executed",
      metrics: config["metrics"] || { "echo" => true }
    )
  end
end
