class Api::V1::Sessions::MagicLinksController < Api::V1::BaseController
  include Authentication::ViaMagicLink

  allow_unauthenticated_access
  disallow_account_scope
  rate_limit to: 10, within: 15.minutes, only: :create, with: :rate_limit_exceeded

  def create
    if email_pending_authentication.blank?
      render_json_unauthorized_message("Enter your email address to sign in.")
    elsif magic_link = MagicLink.consume(code)
      authenticate magic_link
    else
      render_json_unauthorized_message("Try another code.")
    end
  end

  private
    def code
      params.expect(:code)
    end

    def authenticate(magic_link)
      if ActiveSupport::SecurityUtils.secure_compare(email_pending_authentication, magic_link.identity.email)
        sign_in magic_link
      else
        render_json_unauthorized_message("Something went wrong. Please try again.")
      end
    end

    def sign_in(magic_link)
      clear_pending_authentication_token
      session = start_new_session_for(magic_link.identity)

      render_json json: { session_token: session.signed_id }
    end

    def render_json_unauthorized_message(message)
      render_json_error(status: :unauthorized, message: message, code: "UNAUTHORIZED")
    end

    def rate_limit_exceeded
      render_json_too_many_requests
    end
end
