class Identity::AccessToken < ApplicationRecord
  belongs_to :identity

  has_secure_token :token

  enum :permission, %w[ read write ].index_by(&:itself), default: "write"

  validates :description, length: { maximum: 255 }

  def allows?(method)
    case method.to_s.upcase
    when "GET", "HEAD", "OPTIONS"
      true
    else
      write?
    end
  end

  def touch_last_used_at
    touch(:last_used_at)
  end
end
