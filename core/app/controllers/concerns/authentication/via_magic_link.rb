module Authentication::ViaMagicLink
  extend ActiveSupport::Concern

  included do
    after_action :ensure_development_magic_link_not_leaked
  end

  private
    def ensure_development_magic_link_not_leaked
      return unless respond_to?(:flash)

      if !Rails.env.development? && flash[:magic_link_code].present?
        raise "Leaking magic link via flash in #{Rails.env}?"
      end
    end

    def redirect_to_session_magic_link(magic_link, return_to: nil)
      serve_development_magic_link magic_link
      set_pending_authentication_token magic_link
      session[:return_to_after_authenticating] = return_to if return_to

      redirect_to main_app.session_magic_link_url(script_name: nil)
    end

    def serve_development_magic_link(magic_link)
      if Rails.env.development? && magic_link.present?
        flash[:magic_link_code] = magic_link.code if respond_to?(:flash)
        response.set_header("X-Magic-Link-Code", magic_link.code)
      end
    end

    def set_pending_authentication_token(magic_link)
      cookies[:pending_authentication_token] = {
        value: generate_pending_authentication_token(magic_link),
        httponly: true,
        same_site: :lax,
        secure: !Rails.env.local?,
        domain: ENV["SESSION_COOKIE_DOMAIN"].presence,
        expires: magic_link.expires_at
      }.compact
    end

    def pending_authentication_token
      params[:pending_authentication_token].presence || cookies[:pending_authentication_token]
    end

    def email_pending_authentication
      verify_pending_authentication_token(pending_authentication_token)
    end

    def clear_pending_authentication_token
      cookies.delete(:pending_authentication_token, **{ domain: ENV["SESSION_COOKIE_DOMAIN"].presence }.compact)
    end

    def generate_pending_authentication_token(magic_link)
      pending_authentication_token_verifier.generate(
        magic_link.identity.email,
        expires_at: magic_link.expires_at
      )
    end

    def verify_pending_authentication_token(token)
      pending_authentication_token_verifier.verified(token)
    end

    def pending_authentication_token_verifier
      Rails.application.message_verifier(:pending_authentication)
    end
end
