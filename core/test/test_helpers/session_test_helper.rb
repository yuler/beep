module SessionTestHelper
  def sign_in_as(identity, account: nil)
    Current.identity = identity
    Current.account = account || identity.accounts.first
    Current.session = identity.sessions.create!

    ActionDispatch::TestRequest.create.cookie_jar.tap do |cookie_jar|
      cookie_jar.signed[:session_id] = Current.session.signed_id
      cookies[:session_id] = cookie_jar[:session_id]
    end
  end

  def sign_out
    Current.session&.destroy!
    cookies.delete(:session_id)
  end
end

ActiveSupport.on_load(:action_dispatch_integration_test) do
  include SessionTestHelper
end
