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

    # 2. From cookies (PARAGLIDE_LOCALE or locale)
    raw_cookie = cookies[:PARAGLIDE_LOCALE] || cookies[:locale]
    if raw_cookie.present?
      loc = normalize_locale(raw_cookie)
      return loc if loc && I18n.available_locales.include?(loc)
    end

    # 3. From Accept-Language header
    extract_locale_from_accept_language_header
  end

  def normalize_locale(locale_str)
    str = locale_str.to_s.strip.downcase
    return nil if str.blank?

    str.to_sym
  end

  def extract_locale_from_accept_language_header
    header = request.env["HTTP_ACCEPT_LANGUAGE"]
    return nil if header.blank?

    header.scan(/[a-z]{2}(?:-[A-Za-z]{2,})?/i).each do |loc|
      # Match primary language subtag (e.g. "zh-CN" -> :zh, "en-US" -> :en)
      primary = loc.split(/[-_]/).first&.downcase&.to_sym
      return primary if primary && I18n.available_locales.include?(primary)
    end

    nil
  end
end
