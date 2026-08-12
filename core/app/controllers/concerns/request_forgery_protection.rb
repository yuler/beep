module RequestForgeryProtection
  extend ActiveSupport::Concern

  included do
    include ActionController::RequestForgeryProtection

    protect_from_forgery using: :header_only, with: :exception
  end

  private
    def verified_via_header_only?
      super || allowed_api_request?
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
