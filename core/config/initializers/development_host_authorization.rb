# Development: friendly 403 for blocked hosts. Rails's default blocked-host page
# only lists config.hosts entries; this one also tells the developer the canonical
# core URL to use instead (a wrong *.localhost or plain loopback fails fast).
if Rails.env.development?
  class DevelopmentHostAuthorization
    CANONICAL_HOST = "core.#{ENV.fetch("APP_HOST")}"
    CANONICAL_PORT = ENV.fetch("CORE_PORT")
    CANONICAL_URL = "http://#{CANONICAL_HOST}:#{CANONICAL_PORT}"

    def self.call(env)
      blocked_hosts = env["action_dispatch.blocked_hosts"]
      body = if env["action_dispatch.show_detailed_exceptions"]
        render_html(blocked_hosts)
      else
        render_plain(blocked_hosts)
      end

      [ 403, { Rack::CONTENT_TYPE => content_type(env), Rack::CONTENT_LENGTH => body.bytesize.to_s }, [ body ] ]
    end

    def self.content_type(env)
      request = ActionDispatch::Request.new(env)
      if request.xhr?
        "text/plain; charset=#{ActionDispatch::Response.default_charset}"
      else
        "text/html; charset=#{ActionDispatch::Response.default_charset}"
      end
    end

    def self.render_plain(blocked_hosts)
      [
        "Blocked hosts: #{blocked_hosts.join(", ")}",
        "Use #{CANONICAL_URL} instead."
      ].join("\n")
    end

    def self.render_html(blocked_hosts)
      <<~HTML
        <!DOCTYPE html>
        <html lang="en">
        <head>
          <meta charset="utf-8" />
          <meta name="viewport" content="width=device-width, initial-scale=1">
          <title>Blocked host</title>
          <style>
            body { font-family: -apple-system, "Segoe UI", sans-serif; padding: 48px; line-height: 1.5; }
            h1 { font-size: 20px; }
            code { background: #f5f5f5; padding: 2px 6px; border-radius: 4px; font-size: 15px; }
            a { color: #0b57d0; text-decoration: none; font-size: 15px; }
          </style>
        </head>
        <body>
          <h1>Blocked hosts: #{blocked_hosts.join(", ")}</h1>
          <p>This is the <strong>beep</strong> development server. Wrong host or plain
             loopback is refused so a different project&rsquo;s <code>*.localhost</code> can&rsquo;t
             load it. Use the canonical URL:</p>
          <p><a href="#{CANONICAL_URL}">#{CANONICAL_URL}</a></p>
          <p>To allow these hosts, add them to <code>config.hosts</code> in
             <code>core/config/environments/development.rb</code>.</p>
        </body>
        </html>
      HTML
    end
  end

  Rails.application.config.host_authorization = { response_app: DevelopmentHostAuthorization }
end
