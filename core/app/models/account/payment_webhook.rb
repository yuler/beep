class Account::PaymentWebhook < ApplicationRecord
  belongs_to :account
end
