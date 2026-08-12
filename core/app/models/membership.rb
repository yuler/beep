class Membership < ApplicationRecord
  belongs_to :account
  belongs_to :user, optional: true

  enum :role, %w[owner admin member system].index_by(&:itself)
end
