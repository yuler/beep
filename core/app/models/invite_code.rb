class InviteCode < ApplicationRecord
  LENGTH = 6

  before_validation :generate_code, on: :create
  validates :code, uniqueness: true, presence: true

  class << self
    def claim(code)
      if invite_code = find_by(code: Base32.sanitize(code))
        invite_code.destroy!
        true
      end
    end

    def valid?(code)
      find_by(code: Base32.sanitize(code)).present?
    end

    def generate
      create!.code
    end
  end

  private
    def generate_code
      self.code = loop do
        code = Base32.generate(LENGTH)
        break code unless self.class.exists?(code:)
      end
    end
end
