class Api::V1::Test::PrivateController < Api::V1::BaseController
  def show
    render_json json: {
      ok: true,
      slug: Current.account&.slug,
      account_id: Current.account&.id,
      user_id: Current.user&.id,
      identity_id: Current.identity&.id
    }
  end
end
