module Authorization
  extend ActiveSupport::Concern

  included do
    before_action :ensure_can_access_account, if: -> { Current.account.present? && authenticated? }
    after_action :remember_last_account, if: -> { Current.user&.active? }
  end

  class_methods do
    def allow_unauthorized_access(**options)
      skip_before_action :ensure_can_access_account, **options
    end

    def require_access_without_a_user(**options)
      skip_before_action :ensure_can_access_account, **options
      before_action :redirect_existing_user, **options
    end
  end

  private
    def ensure_staff
      head :forbidden unless Current.identity.staff?
    end

    def ensure_admin
      head :forbidden unless Current.user.admin?
    end

    def ensure_can_access_account
      if Current.user.blank? || !Current.user.active?
        # 404 (not 403) to avoid confirming account existence to non-members.
        head :not_found
      end
    end

    def remember_last_account
      return if Current.account.blank?
      # Only when the request targeted an account (path / X-Account-Slug), not the
      # personal fallback used by unscoped API controllers.
      return if request.env["account_slug"].blank?

      write_last_account_slug(Current.account.slug)
    end

    def write_last_account_slug(slug)
      identity = Current.identity
      return if identity.blank?
      return if identity.last_account_slug == slug

      identity.update_column(:last_account_slug, slug)
    end

    def redirect_existing_user
      if Current.user
        redirect_to root_path, alert: "You are already signed in."
      end
    end
end
