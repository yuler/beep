class Account::SlugHold < ApplicationRecord
  HOLD_FOR = 30.days

  belongs_to :account

  validates :slug, presence: true
  validates :expires_at, presence: true

  scope :active, -> { where("expires_at > ?", Time.current) }

  def self.hold!(slug, account:)
    create!(slug: slug, account: account, expires_at: HOLD_FOR.from_now)
  end
end
