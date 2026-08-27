class BeeperApp::Receivers::SslExpiry < BeeperApp::Receivers::Base
  DEFAULT_PORT = 443
  DEFAULT_ALERT_DAYS_BEFORE = 14
  CONNECT_TIMEOUT = 5

  def call
    raw_host = config["hostname"].to_s.strip
    port = (config["port"] || DEFAULT_PORT).to_i
    alert_days_before = (config["alert_days_before"] || DEFAULT_ALERT_DAYS_BEFORE).to_i

    if raw_host.blank?
      return BeeperApp::Signal.new(
        status: :error,
        title: "Configuration error",
        message: "Hostname is required"
      )
    end

    hostname = sanitize_hostname(raw_host)

    resolved_ip = SsrfProtection.resolve_public_ip(hostname)
    if resolved_ip.nil?
      return BeeperApp::Signal.new(
        status: :error,
        title: "Blocked target address",
        message: "Host #{hostname} resolved to a private/disallowed IP address or cannot be resolved"
      )
    end

    cert = fetch_peer_certificate(resolved_ip: resolved_ip, hostname: hostname, port: port)
    not_after = cert.not_after
    seconds_remaining = not_after - Time.current
    days_remaining = (seconds_remaining / 86400.0).floor

    metrics = {
      "days_remaining" => [ days_remaining, 0 ].max,
      "expires_at" => not_after.utc.iso8601
    }

    if days_remaining <= 0
      BeeperApp::Signal.new(
        status: :alerting,
        title: "SSL certificate has expired",
        message: "SSL certificate for #{hostname} expired #{days_remaining.abs} days ago (#{not_after.strftime('%Y-%m-%d %H:%M:%S UTC')})",
        metrics: metrics
      )
    elsif days_remaining < alert_days_before
      BeeperApp::Signal.new(
        status: :alerting,
        title: "SSL certificate expiring soon",
        message: "SSL certificate for #{hostname} expires in #{days_remaining} days (#{not_after.strftime('%Y-%m-%d %H:%M:%S UTC')})",
        metrics: metrics
      )
    else
      BeeperApp::Signal.new(
        status: :ok,
        title: "SSL certificate is valid",
        message: "SSL certificate for #{hostname} is valid for #{days_remaining} more days",
        metrics: metrics
      )
    end
  rescue OpenSSL::SSL::SSLError => e
    BeeperApp::Signal.new(
      status: :alerting,
      title: "SSL handshake failed",
      message: "SSL error for #{hostname || raw_host}: #{e.message}"
    )
  rescue Timeout::Error => e
    BeeperApp::Signal.new(
      status: :alerting,
      title: "SSL signal timed out",
      message: "SSL connection to #{hostname || raw_host} timed out after #{CONNECT_TIMEOUT}s"
    )
  rescue StandardError => e
    BeeperApp::Signal.new(
      status: :alerting,
      title: "SSL signal failed",
      message: "Could not inspect SSL certificate for #{hostname || raw_host}: #{e.message}"
    )
  end

  private

  def sanitize_hostname(value)
    # Strip URL schemes (http://, https://), paths, and port numbers if provided
    cleaned = value.sub(%r{\A[a-zA-Z]+://}, "")
    cleaned = cleaned.split("/").first || ""
    cleaned.split(":").first || ""
  end

  def fetch_peer_certificate(resolved_ip:, hostname:, port:)
    tcp_socket = nil
    ssl_socket = nil
    Timeout.timeout(CONNECT_TIMEOUT) do
      tcp_socket = Socket.tcp(resolved_ip, port, connect_timeout: CONNECT_TIMEOUT)
      ctx = OpenSSL::SSL::SSLContext.new
      ctx.set_params(verify_mode: OpenSSL::SSL::VERIFY_PEER)

      ssl_socket = OpenSSL::SSL::SSLSocket.new(tcp_socket, ctx)
      ssl_socket.hostname = hostname # SNI
      ssl_socket.sync_close = true
      ssl_socket.connect

      cert = ssl_socket.peer_cert
      ssl_socket.close
      cert
    end
  ensure
    ssl_socket&.close rescue nil
    tcp_socket&.close rescue nil
  end
end
