# frozen_string_literal: true

module SetLocale
  extend ActiveSupport::Concern

  included do
    around_action :switch_locale
  end

  private

  def switch_locale(&action)
    locale = extract_locale || I18n.default_locale
    I18n.with_locale(locale, &action)
  end

  def extract_locale
    # 1. From params[:locale]
    if params[:locale].present?
      loc = normalize_locale(params[:locale])
      return loc if loc && I18n.available_locales.include?(loc)
    end

    # 2. From cookies[:locale]
    if cookies[:locale].present?
      loc = normalize_locale(cookies[:locale])
      return loc if loc && I18n.available_locales.include?(loc)
    end

    # 3. From Accept-Language header
    extract_locale_from_accept_language_header
  end

  def normalize_locale(locale_str)
    str = locale_str.to_s.strip
    return nil if str.blank?

    # Backward compatibility: map zh-CN / zh_CN / zh-* to :zh
    return :zh if str.casecmp?("zh-CN") || str.casecmp?("zh_CN") || str.casecmp?("zh")

    str.to_sym
  end

  def extract_locale_from_accept_language_header
    header = request.env["HTTP_ACCEPT_LANGUAGE"]
    return nil if header.blank?

    header.scan(/[a-z]{2}(?:-[A-Z]{2})?/i).each do |loc|
      normalized = loc.tr("_", "-")

      # Special handling for Chinese variations
      if normalized.downcase.start_with?("zh") && I18n.available_locales.include?(:zh)
        return :zh
      end

      matched = I18n.available_locales.find { |l| l.to_s.casecmp?(normalized) }
      return matched if matched
    end

    nil
  end
end
