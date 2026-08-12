module ApiAuthentication
  extend ActiveSupport::Concern

  included do
    # Auth before account lookup so unknown vs known slugs both 401 when unauthenticated.
    before_action :require_authentication
    before_action :require_account
    before_action :ensure_account_user
    helper_method :authenticated?

    etag { Current.identity.id if authenticated? }

    include ActionController::HttpAuthentication::Token::ControllerMethods
  end

  class_methods do
    def skip_account_scope(**options)
      skip_before_action :require_account, **options
      skip_before_action :ensure_account_user, **options
    end

    def disallow_account_scope(**options)
      skip_account_scope(**options)
      before_action :reject_account_scoped_request, **options
    end

    def require_unauthenticated_access(**options)
      allow_unauthenticated_access(**options)
      before_action :redirect_authenticated_user, **options
    end

    def allow_unauthenticated_access(**options)
      skip_before_action :require_authentication, **options
      before_action :resume_session, **options
      allow_unauthorized_access(**options)
    end
  end

  private
    def authenticated?
      Current.identity.present?
    end

    def require_authentication
      resume_session ||
        authenticate_by_bearer_token ||
        authenticate_by_query_token ||
        json_request_unauthorized
    end

    def resume_session
      if session = find_session_by_cookie
        Current.session = session
      end
    end

    def find_session_by_cookie
      Session.find_signed(cookies.signed[:session_id])
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

    def authenticate_by_bearer_token
      if request.authorization.to_s.include?("Bearer")
        if bearer_token_authenticatable_request?
          authenticate_or_request_with_http_token do |token|
            if session = Session.find_signed(token)
              Current.session = session
            elsif identity = Identity.find_by_permissable_access_token(token, method: request.method)
              Current.identity = identity
            end
          end
        else
          request_http_token_authentication
        end
      end
    end

    def bearer_token_authenticatable_request?
      request.format.json?
    end

    # NOTE: Query parameter tokens are convenient for testing but can be leaked via
    # server logs and browser history. Consider disabling this in production.
    def authenticate_by_query_token
      if token = params[:token]
        if session = Session.find_signed(token)
          Current.session = session
        elsif identity = Identity.find_by_permissable_access_token(token, method: request.method)
          Current.identity = identity
        end
      end
    end

    def json_request_unauthorized
      render json: { error: "Unauthorized" }, status: :unauthorized
    end

    # Slug/header already resolved by middleware into Current.account when present.
    # Unscoped requests are allowed through so ensure_account_user can fall back to personal.
    def require_account
      return if request.env["account_slug"].blank?

      json_request_account_not_found if Current.account.nil?
    end

    def ensure_account_user
      if Current.account.present?
        if (user = Current.identity.users.find_by(account: Current.account))
          Current.user = user
        else
          json_request_account_not_found
        end
      else
        personal_account = Current.identity.personal_account
        Current.account = personal_account
        Current.user = Current.identity.users.find_by(account: personal_account)
      end
    end

    def reject_account_scoped_request
      if Current.account.present? || request.env["account_slug"].present?
        render json: { error: "Account scope not allowed" }, status: :not_found
      end
    end

    def json_request_account_not_found
      render json: { error: "Account not found" }, status: :not_found
    end
end
