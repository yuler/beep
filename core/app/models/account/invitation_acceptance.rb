class Account::InvitationAcceptance < ApplicationRecord
  belongs_to :invitation, class_name: "Account::Invitation"
  belongs_to :identity
  belongs_to :user

  validates :invitation_id, uniqueness: true
end
