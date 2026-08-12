class Base32
  LOOKALIKE_CHARS = { "O" => "0", "I" => "1", "L" => "1" }.freeze
  VALID_CHARS_PATTERN = /[^#{SecureRandom::BASE32_ALPHABET.join}]/

  class << self
    def generate(length = 6)
      SecureRandom.base32(length)
    end

    def sanitize(input)
      if input.blank?
        return
      end

      normalized = input.to_s.upcase
      # Substitute lookalike characters
      substituted = LOOKALIKE_CHARS.each_with_object(normalized) { |(from, to), str| str.gsub!(from, to) || str }
      # Remove invalid characters
      substituted.gsub(VALID_CHARS_PATTERN, "")
    end
  end
end
