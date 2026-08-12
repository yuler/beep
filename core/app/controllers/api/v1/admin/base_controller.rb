class Api::V1::Admin::BaseController < Api::V1::BaseController
  disallow_account_scope
  before_action :require_staff

  private
    def require_staff
      unless Current.identity.staff?
        render_json_error(status: :forbidden, message: "Forbidden", code: "FORBIDDEN")
      end
    end
end
