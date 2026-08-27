class Beeper::Receivers::SiteUptime < Beeper::Receivers::Base
  MAX_REDIRECTS = 3
  MAX_BODY_BYTES = 8.kilobytes
  DEFAULT_TIMEOUT_MS = 3000

  def call
    target_url = config["target_url"].to_s.strip
    expected_status = (config["expected_status"] || 200).to_i
    timeout_ms = [ (config["timeout_ms"] || DEFAULT_TIMEOUT_MS).to_i, 10_000 ].min
    timeout_seconds = timeout_ms / 1000.0

    if target_url.blank?
      return Beeper::Signal.new(
        status: :error,
        title: "Configuration error",
        message: "Target URL is required"
      )
    end

    uri = URI.parse(target_url)
    unless uri.is_a?(URI::HTTP) || uri.is_a?(URI::HTTPS)
      return Beeper::Signal.new(
        status: :error,
        title: "Invalid URL",
        message: "Target URL must start with http:// or https://"
      )
    end

    start_time = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    response, final_uri = fetch_with_redirects(uri, timeout: timeout_seconds, redirects_remaining: MAX_REDIRECTS)
    elapsed_ms = ((Process.clock_gettime(Process::CLOCK_MONOTONIC) - start_time) * 1000).round

    status_code = response.code.to_i
    metrics = {
      "status" => status_code,
      "latency_ms" => elapsed_ms
    }

    if status_code == expected_status
      Beeper::Signal.new(
        status: :ok,
        title: "Site is operational",
        message: "HTTP #{status_code} from #{final_uri.host} (#{elapsed_ms}ms)",
        metrics: metrics
      )
    else
      Beeper::Signal.new(
        status: :alerting,
        title: "Site returned HTTP #{status_code}",
        message: "Expected HTTP #{expected_status} but received HTTP #{status_code} from #{final_uri.host} (#{elapsed_ms}ms)",
        metrics: metrics
      )
    end
  rescue SsrfProtection::BlockedAddressError => e
    Beeper::Signal.new(
      status: :error,
      title: "Blocked target address",
      message: e.message
    )
  rescue Net::OpenTimeout, Net::ReadTimeout, Timeout::Error => e
    Beeper::Signal.new(
      status: :alerting,
      title: "Site signal timed out",
      message: "Connection to #{uri&.host || target_url} timed out after #{timeout_ms}ms"
    )
  rescue StandardError => e
    Beeper::Signal.new(
      status: :alerting,
      title: "Site signal failed",
      message: "Failed to connect to #{uri&.host || target_url}: #{e.message}"
    )
  end

  private

  def fetch_with_redirects(uri, timeout:, redirects_remaining:)
    if redirects_remaining < 0
      raise StandardError, "Too many HTTP redirects (exceeded limit of #{MAX_REDIRECTS})"
    end

    resolved_ip = SsrfProtection.resolve_public_ip(uri.host)
    if resolved_ip.nil?
      raise SsrfProtection::BlockedAddressError, "Host #{uri.host} resolved to a private/disallowed IP address or cannot be resolved"
    end

    http = Net::HTTP.new(uri.host, uri.port)
    http.ipaddr = resolved_ip
    http.open_timeout = timeout
    http.read_timeout = timeout

    if uri.scheme == "https"
      http.use_ssl = true
      http.verify_mode = OpenSSL::SSL::VERIFY_PEER
    end

    request = Net::HTTP::Get.new(uri.request_uri)
    request["Host"] = uri.host
    request["User-Agent"] = "Beep-Signal-Uptime/1.0"

    response = http.request(request)

    if response.is_a?(Net::HTTPRedirection) && response["location"].present?
      new_uri = URI.join(uri.to_s, response["location"])
      unless new_uri.is_a?(URI::HTTP) || new_uri.is_a?(URI::HTTPS)
        raise StandardError, "Redirect destination must be HTTP/HTTPS"
      end
      fetch_with_redirects(new_uri, timeout: timeout, redirects_remaining: redirects_remaining - 1)
    else
      [ response, uri ]
    end
  end
end
