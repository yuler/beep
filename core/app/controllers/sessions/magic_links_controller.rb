class Sessions::MagicLinksController < ApplicationController
  disallow_account_scope
  require_unauthenticated_access
  rate_limit to: 10, within: 15.minutes, only: :create, with: :rate_limit_exceeded
  before_action :ensure_that_email_pending_authentication_exists

  # TODO: layout
  # layout "public"

  def show
  end

  def create
    if magic_link = MagicLink.consume(code)
      authenticate magic_link
    else
      invalid_code
    end
  end

  private
    def ensure_that_email_pending_authentication_exists
      unless email_pending_authentication.present?
        redirect_to new_session_path, alert: "Enter your email address to sign in."
      end
    end

    def code
      params.expect(:code)
    end

    def authenticate(magic_link)
      if ActiveSupport::SecurityUtils.secure_compare(email_pending_authentication, magic_link.identity.email)
        sign_in magic_link
      else
        email_mismatch
      end
    end

    def sign_in(magic_link)
      clear_pending_authentication_token
      start_new_session_for magic_link.identity
      redirect_to after_sign_in_url(magic_link)
    end

    def after_sign_in_url(magic_link)
      after_authentication_url
    end

    def email_mismatch
      clear_pending_authentication_token
      redirect_to new_session_path, alert: "Something went wrong. Please try again."
    end

    def invalid_code
      redirect_to session_magic_link_path, flash: { shake: true }
    end

    def rate_limit_exceeded
      redirect_to session_magic_link_path, alert: "Try again in 15 minutes."
    end
end
