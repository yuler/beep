module Authentication
  extend ActiveSupport::Concern

  included do
    before_action :require_account # Checking account must happen first
    before_action :require_authentication
    helper_method :authenticated?

    # before_action: set_sentry_context, TODO:

    include Authentication::ViaMagicLink
  end

  class_methods do
    def require_unauthenticated_access(**options)
      allow_unauthenticated_access(**options)
      before_action :redirect_authenticated_user, **options
    end

    def allow_unauthenticated_access(**options)
      skip_before_action :require_authentication, **options
      before_action :resume_session, **options
      allow_unauthorized_access(**options)
    end

    alias_method :skip_authentication, :allow_unauthenticated_access

    def disallow_account_scope(**options)
      skip_before_action :require_account, **options
      before_action :redirect_account_scoped_request, **options
    end
  end

  private
    def authenticated?
      Current.identity.present?
    end

    def require_account
      if !Current.account.present?
        redirect_to main_app.my_accounts_url(script_name: nil)
      end
    end

    def require_authentication
      resume_session || authenticate_by_bearer_token || request_authentication
    end

    def resume_session
      if session = find_session_by_cookie
        set_current_session session
      end
    end

    def find_session_by_cookie
      Session.find_signed(cookies.signed[:session_id])
    end

    def authenticate_by_bearer_token
      if request.authorization.to_s.include?("Bearer")
        authenticate_or_request_with_http_token do |token|
          if identity = Identity.find_by_permissable_access_token(token, method: request.method)
            Current.identity = identity
          end
        end
      end
    end

    def request_authentication
      session[:return_to_after_authenticating] = request.url
      redirect_to main_app.new_session_path(script_name: nil)
    end

    def after_authentication_url
      return_to = session.delete(:return_to_after_authenticating)

      if return_to.present?
        return_to
      else
        accounts = Current.identity.accounts
        if accounts.one?
          main_app.root_url(script_name: accounts.first.slug_path)
        else
          main_app.my_accounts_url(script_name: nil)
        end
      end
    end

    def redirect_authenticated_user
      if authenticated?
        redirect_to main_app.root_url, notice: "You are already signed in."
      end
    end

    def redirect_account_scoped_request
      redirect_to main_app.root_url(script_name: nil) if Current.account.present?
    end

    def start_new_session_for(identity)
      identity.sessions.create!(user_agent: request.user_agent, ip_address: request.remote_ip).tap do |session|
        set_current_session session
      end
    end

    def set_current_session(session)
      Current.session = session
      cookies.signed.permanent[:session_id] = {
        value: session.signed_id,
        httponly: true,
        same_site: :lax,
        secure: !Rails.env.local?,
        domain: ENV["SESSION_COOKIE_DOMAIN"].presence
      }.compact
    end

    def terminate_session
      Current.session&.destroy
      cookies.delete(:session_id, **{ domain: ENV["SESSION_COOKIE_DOMAIN"].presence }.compact)
    end
end
