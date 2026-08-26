class Plugin::Checkers::SslExpiry < Plugin::Checkers::Base
  DEFAULT_PORT = 443
  DEFAULT_ALERT_DAYS_BEFORE = 14
  CONNECT_TIMEOUT = 5

  def call
    hostname = config["hostname"].to_s.strip
    port = (config["port"] || DEFAULT_PORT).to_i
    alert_days_before = (config["alert_days_before"] || DEFAULT_ALERT_DAYS_BEFORE).to_i

    if hostname.blank?
      return Plugin::CheckResult.new(
        status: :error,
        title: "Configuration error",
        message: "Hostname is required"
      )
    end

    resolved_ip = SsrfProtection.resolve_public_ip(hostname)
    if resolved_ip.nil?
      return Plugin::CheckResult.new(
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
      Plugin::CheckResult.new(
        status: :alerting,
        title: "SSL certificate has expired",
        message: "SSL certificate for #{hostname} expired #{days_remaining.abs} days ago (#{not_after.strftime('%Y-%m-%d %H:%M:%S UTC')})",
        metrics: metrics
      )
    elsif days_remaining < alert_days_before
      Plugin::CheckResult.new(
        status: :alerting,
        title: "SSL certificate expiring soon",
        message: "SSL certificate for #{hostname} expires in #{days_remaining} days (#{not_after.strftime('%Y-%m-%d %H:%M:%S UTC')})",
        metrics: metrics
      )
    else
      Plugin::CheckResult.new(
        status: :ok,
        title: "SSL certificate is valid",
        message: "SSL certificate for #{hostname} is valid for #{days_remaining} more days",
        metrics: metrics
      )
    end
  rescue OpenSSL::SSL::SSLError => e
    Plugin::CheckResult.new(
      status: :alerting,
      title: "SSL handshake failed",
      message: "SSL error for #{hostname}: #{e.message}"
    )
  rescue StandardError => e
    Plugin::CheckResult.new(
      status: :alerting,
      title: "SSL check failed",
      message: "Could not inspect SSL certificate for #{hostname}: #{e.message}"
    )
  end

  private

  def fetch_peer_certificate(resolved_ip:, hostname:, port:)
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
  ensure
    ssl_socket&.close rescue nil
    tcp_socket&.close rescue nil
  end
end
