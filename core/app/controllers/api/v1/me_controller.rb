class Api::V1::MeController < Api::V1::BaseController
  disallow_account_scope

  def show
    render_json json: {
      identity: {
        id: Current.identity.id,
        email: Current.identity.email,
        name: Current.identity.full_name,
        staff: Current.identity.staff?
      },
      accounts: Current.identity.accounts.order(:name).map { |account|
        {
          id: account.id,
          name: account.name,
          slug: account.slug,
          personal: account.personal?
        }
      },
      last_account_slug: Current.identity.last_account_slug.presence
    }
  end

  def update_last_account
    account = Current.identity.accounts.find_by(slug: params.require(:slug))
    unless account
      return render_json_error(status: :not_found, message: "Account not found", code: "NOT_FOUND")
    end

    write_last_account_slug(account.slug)
    render_json json: { last_account_slug: account.slug }
  end
end
