# Development-only CORS for apps/web calling /api/v1 from the canonical web host.
if Rails.env.development?
  class DevelopmentCors
    API_PREFIX = "/api/v1"
    # Align with config.hosts: only the canonical web host (from APP_HOST) is an
    # allowed origin. A wrong *.localhost origin is refused.
    WEB_HOST = ENV.fetch("APP_HOST", "beep.localhost")
    WEB_ORIGIN = %r{\Ahttps?://web\.#{Regexp.escape(WEB_HOST)}(:\d+)?\z}i

    def initialize(app)
      @app = app
    end

    def call(env)
      request = Rack::Request.new(env)

      if api_request?(request)
        origin = env["HTTP_ORIGIN"]

        if request.request_method == "OPTIONS"
          return preflight_response(origin)
        end

        status, headers, body = @app.call(env)
        return [ status, cors_headers(headers, origin), body ]
      end

      @app.call(env)
    end

    private
      def api_request?(request)
        request.path.start_with?(API_PREFIX)
      end

      def allowed_origin?(origin)
        return false if origin.blank?

        origin.match?(WEB_ORIGIN)
      end

      def preflight_response(origin)
        [
          204,
          cors_headers({
            "Access-Control-Allow-Methods" => "GET, POST, PUT, PATCH, DELETE, OPTIONS",
            "Access-Control-Allow-Headers" => "Authorization, Content-Type, X-Account-Slug",
            "Access-Control-Expose-Headers" => "X-Magic-Link-Code",
            "Access-Control-Max-Age" => "86400"
          }, origin),
          []
        ]
      end

      def cors_headers(headers, origin)
        headers = headers.dup
        if allowed_origin?(origin)
          headers["Access-Control-Allow-Origin"] = origin
          headers["Access-Control-Allow-Credentials"] = "true"
          headers["Access-Control-Expose-Headers"] = [
            headers["Access-Control-Expose-Headers"],
            "X-Magic-Link-Code"
          ].compact.join(", ")
          headers["Vary"] = [ headers["Vary"], "Origin" ].compact.join(", ")
        end
        headers
      end
  end

  Rails.application.config.middleware.insert_before 0, DevelopmentCors
end
