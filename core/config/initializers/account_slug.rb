module AccountSlug
  PATTERN = /([a-zA-Z0-9_-]{4,16})/
  FORMAT = /\A[a-z0-9\-_]+\z/i
  LENGTH = 4..16
  # Require the slug segment to end at `/` or EOS so longer prefixes (e.g. invitations) never partial-match.
  PATH_INFO_MATCH = /\A(\/#{AccountSlug::PATTERN})(?=\/|\z)/

  # Must never be confused with an account slug. Two groups:
  # - FROM_ROUTES: top-level segments in config/routes.rb (+ mounts). Update when adding routes.
  # - EXTRA: policy / future / infra words reserved even without a matching route today.
  RESERVED_FROM_ROUTES = %w[
    admin api cable home hotwire-spark invitations join join_code
    beeper_apps beepers landing letter_opener manifest my payment rails runner runners service-worker session settings
    subscription up users webhooks
  ].freeze
  RESERVED_EXTRA = %w[
    assets billing dev device help jobs landings login logout magic_link setup static
    support test web_push
  ].freeze
  RESERVED_SLUGS = (RESERVED_FROM_ROUTES + RESERVED_EXTRA).freeze

  ACCOUNT_SLUG_HEADER = "HTTP_X_ACCOUNT_SLUG"

  class Extractor
    def initialize(app)
      @app = app
    end

    # We're using account slug prefixes in the URL path. Rather than namespace
    # all our routes, we "mount" the Rails app at that prefix for web requests.
    # API: /api/vN/{slug}/... (path wins over X-Account-Slug header).
    def call(env)
      request = ActionDispatch::Request.new(env)
      extract_account_slug!(request, env)

      if env["account_slug"]
        account = Account.find_by(slug: env["account_slug"])
        dispatch_account_slug(request, env, account)
      else
        Current.without_account do
          @app.call env
        end
      end
    end

    private
      def dispatch_account_slug(request, env, account)
        if account
          Current.with_account(account) { @app.call env }
        elsif api_request?(request)
          # Defer missing-slug 404 until after API auth so unknown vs known
          # unauthenticated requests both return 401.
          Current.without_account { @app.call env }
        elsif signed_in?(request)
          not_found(env)
        else
          # Match known-slug unauthenticated response (302 → login) to avoid
          # revealing whether the slug exists.
          redirect_to_login
        end
      end

      def api_request?(request)
        request.path_info.start_with?("/api") || request.script_name.to_s.start_with?("/api")
      end

      def signed_in?(request)
        Session.find_signed(request.cookie_jar.signed[:session_id]).present?
      end

      def redirect_to_login
        [ 302, { "Location" => "/session/new", "Content-Type" => "text/html" }, [ "Redirecting..." ] ]
      end

      def extract_account_slug!(request, env)
        # API: /api/vN/:slug/... → strip slug; path takes priority over header
        if request.path_info =~ %r{\A(/api/v\d+)/(#{AccountSlug::PATTERN})(?=/|\z)} && !$2.in?(RESERVED_SLUGS)
          env["account_slug"] = AccountSlug.decode($2)
          remainder = $'
          request.path_info = remainder.empty? ? $1 : "#{$1}#{remainder}"
        elsif request.path_info.start_with?("/api")
          header = env[ACCOUNT_SLUG_HEADER].to_s
          env["account_slug"] = header.present? && !header.in?(RESERVED_SLUGS) ? AccountSlug.decode(header) : nil
        # Action Cable reconnect: SCRIPT_NAME already carries the slug
        elsif request.script_name.present? && request.script_name =~ PATH_INFO_MATCH && !$2.in?(RESERVED_SLUGS)
          env["account_slug"] = AccountSlug.decode($2)
        # Web: /<slug>/... → move the slug to SCRIPT_NAME so all routes mount under it
        elsif request.path_info =~ PATH_INFO_MATCH && !$2.in?(RESERVED_SLUGS)
          request.engine_script_name = request.script_name = $1
          request.path_info = $'.empty? ? "/" : $'
          env["account_slug"] = AccountSlug.decode($2)
        end
      end

      # Same static / negotiated body Rails uses for missing routes (public/404.html).
      def not_found(env)
        ActionDispatch::PublicExceptions.new(Rails.public_path).call(
          env.merge(
            "PATH_INFO" => "/404",
            "REQUEST_METHOD" => "GET",
            "action_dispatch.original_request_method" => env["REQUEST_METHOD"]
          )
        )
      end
  end

  # Decode slug from URL form to the lookup key.
  def self.decode(slug) slug.to_s end
  # Encode an Account#slug into the SCRIPT_NAME prefix used by view helpers.
  # Example: `root_path(script_name: account.slug_path)` → "/<slug>"
  def self.encode(slug) "/#{slug}" end
end

Rails.application.config.middleware.insert_after Rack::TempfileReaper, AccountSlug::Extractor
