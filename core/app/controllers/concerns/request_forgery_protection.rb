module RequestForgeryProtection
  extend ActiveSupport::Concern

  # NOTE: included at module level (not inside `included do`) on purpose.
  # `base.include` from an `included` hook inserts the module *above* this
  # concern in the ancestry chain, which would silently shadow the
  # `verified_request?` / `verified_via_header_only?` overrides below.
  include ActionController::RequestForgeryProtection

  included do
    protect_from_forgery using: :header_only, with: :exception
  end

  private
    def verified_request?
      return super unless protect_against_forgery?
      return false if cookie_session_mutation_without_xhr_header?

      super
    end

    def verified_via_header_only?
      super || allowed_api_request?
    end

    # Defense in depth on top of SameSite=Lax cookies and Sec-Fetch checks:
    # a request carrying the session cookie was made by a browser, and
    # browsers can only attach `X-Requested-With` from same-origin/proxied
    # JS (the web client sends it on every call) — never from a cross-site
    # form or navigation. Bearer-token and other non-browser API clients
    # carry no cookie and are unaffected.
    def cookie_session_mutation_without_xhr_header?
      !request.get? && !request.head? && !verified_query_request? &&
        request.format.json? &&
        cookies.signed[:session_id].present? &&
        !request.xhr?
    end

    # Allow non-browser API clients (curl, CLI, mobile apps, server-to-server).
    # Browsers always attach `Sec-Fetch-Site`, so its absence means the
    # request did not come from a browser page — and a cross-site request
    # forgery necessarily does.
    #
    # Only `Sec-Fetch-Site` is used as the signal on purpose. The Nitro
    # `/api` proxy in front of core in production (Mode B) normalizes
    # `Accept` to `*/*` (so `request.format.json?` is false) and injects
    # `Sec-Fetch-Mode: cors`, so neither of those can tell proxied
    # non-browser clients apart from browsers. Cookie-carrying browser
    # mutations stay protected by
    # `cookie_session_mutation_without_xhr_header?` above, which runs first.
    def allowed_api_request?
      sec_fetch_site_value.nil?
    end
end
