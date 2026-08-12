class MagicLink < ApplicationRecord
  CODE_LENGTH = 6
  EXPIRATION_TIME = 20.minutes

  belongs_to :identity

  enum :purpose, %w[ sign_in sign_up ], prefix: :for, default: :sign_in

  before_validation :generate_code, on: :create
  before_validation :set_expiration, on: :create

  validates :code, uniqueness: true, presence: true

  scope :active, -> { where(expires_at: Time.current...) }
  scope :stale, -> { where(expires_at: ..Time.current) }

  class << self
    def consume(code)
      active.find_by(code: Base32.sanitize(code))&.consume
    end

    def cleanup
      stale.delete_all
    end
  end

  def consume
    destroy
    self
  end

  private
    def generate_code
      self.code = loop do
        code = Base32.generate(CODE_LENGTH)
        break code unless self.class.exists?(code:)
      end
    end

    def set_expiration
      self.expires_at ||= EXPIRATION_TIME.from_now
    end
end
