class Account::Subscription < ApplicationRecord
  belongs_to :account
  has_many :charges, class_name: "Account::Charge", dependent: :destroy

  enum :status, %w[ active past_due unpaid canceled incomplete incomplete_expired trialing paused ].index_by(&:itself)
end
