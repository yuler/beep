module ApplicationCable
  class Connection < ActionCable::Connection::Base
    identified_by :current_user

    def connect
      set_current_user || reject_unauthorized_connection
    end

    private
      def set_current_user
        if session = find_session_by_cookie
          account = Account.find_by(slug: request.env["account_slug"])

          if account && (user = session.identity.users.find_by(account: account))&.active?
            Current.account = account
            self.current_user = user
          end
        end
      end

      def find_session_by_cookie
        Session.find_signed(cookies.signed[:session_id])
      end
  end
end
