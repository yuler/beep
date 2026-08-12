class Api::V1::Admin::StatsController < Api::V1::Admin::BaseController
  def show
    render_json json: {
      accounts: {
        total: Account.count,
        last_7_days: Account.where(created_at: 7.days.ago..).count,
        last_24_hours: Account.where(created_at: 24.hours.ago..).count
      },
      identities: {
        total: Identity.count,
        last_7_days: Identity.where(created_at: 7.days.ago..).count,
        last_24_hours: Identity.where(created_at: 24.hours.ago..).count
      },
      recent_accounts: Account.order(created_at: :desc).limit(10).map { |account|
        {
          id: account.id,
          name: account.name,
          slug: account.slug,
          personal: account.personal?,
          created_at: account.created_at.iso8601
        }
      }
    }
  end
end
