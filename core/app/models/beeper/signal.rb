Beeper::Signal = Data.define(:status, :title, :message, :metrics) do
  def initialize(status:, title: nil, message: nil, metrics: {})
    status_sym = status.to_sym
    unless [ :ok, :alerting, :error ].include?(status_sym)
      raise ArgumentError, "status must be :ok, :alerting, or :error (got #{status.inspect})"
    end

    super(status: status_sym, title: title, message: message, metrics: metrics || {})
  end

  def ok?
    status == :ok || status == "ok"
  end

  def alerting?
    status == :alerting || status == "alerting"
  end

  def error?
    status == :error || status == "error"
  end

  def to_h
    {
      "status" => status.to_s,
      "title" => title,
      "message" => message,
      "metrics" => metrics
    }.compact
  end
end
