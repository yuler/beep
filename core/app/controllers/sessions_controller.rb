class SessionsController < ApplicationController
  disallow_account_scope
  require_unauthenticated_access except: :destroy
  rate_limit to: 10, within: 3.minutes, only: :create, with: -> { redirect_to new_session_path, alert: "Try again later." }

  def new
  end

  def create
    if identity = Identity.find_by(email: email)
      sign_in identity
    else
      sign_up
    end
  end

  def destroy
    return_to = safe_return_to(params[:return_to])
    email = params[:email]
    terminate_session

    if return_to.present?
      session[:return_to_after_authenticating] = return_to
      redirect_to new_session_path(script_name: nil, email: email), status: :see_other
    else
      redirect_to root_path, status: :see_other
    end
  end

  private
    def email
      params.expect(:email)
    end

    def sign_in(identity)
      redirect_to_session_magic_link identity.send_magic_link(for: :sign_in)
    end

    def sign_up
      signup = Signup.new(email: email)

      if signup.valid?(:identity_creation)
        magic_link = signup.create_identity
        redirect_to_session_magic_link magic_link
      else
        redirect_to new_session_path, alert: signup.errors.full_messages.to_sentence
      end
    end

    def safe_return_to(value)
      return if value.blank?

      uri = URI.parse(value)
      value if uri.scheme.blank? && uri.host.blank? && value.start_with?("/")
    rescue URI::InvalidURIError
      nil
    end
end
