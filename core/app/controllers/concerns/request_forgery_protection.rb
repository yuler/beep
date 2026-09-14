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

    # Allow non-browser API clients (curl, mobile apps, etc.) that never send
    # Sec-Fetch-* headers. Browsers always attach Sec-Fetch-Mode, so its
    # presence means the request is browser-originated and must go through
    # normal CSRF verification regardless of Sec-Fetch-Site.
    def allowed_api_request?
      request.format.json? &&
        sec_fetch_site_value.nil? &&
        request.headers["Sec-Fetch-Mode"].nil?
    end
end
