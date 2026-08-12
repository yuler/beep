class User < ApplicationRecord
  include Role

  belongs_to :account
  belongs_to :identity, optional: true

  validates :name, presence: true

  # TODO: deactivate user
  def deactivate
    transaction do
      # accesses.destroy_all
      # update! active: false, identity: nil
      # close_remote_connections
    end
  end

  def verified?
    verified_at.present?
  end

  def verify
    update!(verified_at: Time.current) unless verified?
  end
end
