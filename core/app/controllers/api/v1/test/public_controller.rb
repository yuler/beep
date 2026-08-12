class Api::V1::Test::PublicController < Api::V1::BaseController
  allow_unauthenticated_access
  skip_account_scope

  def show
    render_json json: { ok: true, slug: Current.account&.slug }
  end
end
