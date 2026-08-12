class Api::V1::SessionsController < Api::V1::BaseController
  include Authentication::ViaMagicLink

  allow_unauthenticated_access only: :create
  disallow_account_scope
  rate_limit to: 10, within: 3.minutes, only: :create, with: :rate_limit_exceeded

  def create
    if identity = Identity.find_by(email: email)
      start_magic_link_for identity.send_magic_link(for: :sign_in)
    else
      sign_up
    end
  end

  def destroy
    terminate_session
    render_json_ok
  end

  private
    def email
      params.expect(:email)
    end

    def sign_up
      signup = Signup.new(email: email)

      if signup.valid?(:identity_creation)
        start_magic_link_for signup.create_identity
      else
        render_json_error(
          status: :unprocessable_entity,
          message: signup.errors.full_messages.to_sentence,
          code: "INVALID_EMAIL"
        )
      end
    end

    def start_magic_link_for(magic_link)
      # Same as HTML: pending auth cookie + (dev) X-Magic-Link-Code — not OTP in JSON.
      serve_development_magic_link(magic_link)
      set_pending_authentication_token(magic_link)

      render_json json: {
        # Non-browser clients; browsers rely on the pending_authentication_token cookie.
        pending_authentication_token: generate_pending_authentication_token(magic_link)
      }
    end

    def rate_limit_exceeded
      render_json_too_many_requests
    end
end
