# Custom UUID attribute type with base36 string representation
# refs: https://github.com/basecamp/fizzy/blob/49c4f2adc6069d8e58f3091a797e9182d85ebbb6/lib/rails_ext/active_record_uuid_type.rb

# Shared base36 encoding utilities for UUID
module UuidBase36
  BASE36_LENGTH = 25 # 36^25 > 2^128

  module_function

  def generate
    uuid = SecureRandom.uuid_v7
    hex = uuid.delete("-")
    normalize(hex.to_i(16))
  end

  def normalize(integer)
    integer.to_s(36).rjust(BASE36_LENGTH, "0")
  end

  def to_hex(base36_value)
    base36_value.to_s.to_i(36).to_s(16).rjust(32, "0")
  end

  def from_hex(hex)
    normalize(hex.to_i(16))
  end
end

module ActiveRecord
  module Type
    # UUID type for SQLite (stored as binary, displayed as base36)
    class Uuid < Binary
      def self.generate
        UuidBase36.generate
      end

      def serialize(value)
        return unless value

        hex = UuidBase36.to_hex(value)
        binary = hex.scan(/../).map(&:hex).pack("C*")
        super(binary)
      end

      def deserialize(value)
        return unless value

        hex = value.to_s.unpack1("H*")
        UuidBase36.from_hex(hex)
      end

      def cast(value)
        value
      end
    end
  end
end

# Register the UUID type for SQLite adapter
ActiveRecord::Type.register(:uuid, ActiveRecord::Type::Uuid, adapter: :sqlite3)
