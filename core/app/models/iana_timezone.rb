module IanaTimezone
  DEFAULT = "UTC"

  class << self
    def valid?(name)
      return false if name.blank?

      TZInfo::Timezone.get(name.to_s)
      true
    rescue TZInfo::InvalidTimezoneIdentifier
      false
    end

    def resolve(preferred, fallback = nil)
      if valid?(preferred)
        preferred.to_s
      elsif valid?(fallback)
        fallback.to_s
      else
        DEFAULT
      end
    end
  end
end
