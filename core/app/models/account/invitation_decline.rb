class Account::InvitationDecline < ApplicationRecord
  belongs_to :invitation, class_name: "Account::Invitation"
  belongs_to :identity

  validates :invitation_id, uniqueness: true
end
