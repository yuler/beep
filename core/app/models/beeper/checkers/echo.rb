class Beeper::Checkers::Echo < Beeper::Checkers::Base
  def call
    status = config["status"] || "ok"
    Beeper::CheckResult.new(
      status: status,
      title: config["title"] || "Echo Check",
      message: config["message"] || "Echo check executed",
      metrics: config["metrics"] || { "echo" => true }
    )
  end
end
