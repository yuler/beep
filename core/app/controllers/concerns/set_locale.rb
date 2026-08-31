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
    if params[:locale].present? && I18n.available_locales.map(&:to_s).include?(params[:locale].to_s)
      return params[:locale].to_sym
    end

    # 2. From cookies[:locale]
    if cookies[:locale].present? && I18n.available_locales.map(&:to_s).include?(cookies[:locale].to_s)
      return cookies[:locale].to_sym
    end

    # 3. From Accept-Language header
    extract_locale_from_accept_language_header
  end

  def extract_locale_from_accept_language_header
    header = request.env["HTTP_ACCEPT_LANGUAGE"]
    return nil if header.blank?

    header.scan(/[a-z]{2}(?:-[A-Z]{2})?/i).each do |loc|
      normalized = loc.tr("_", "-")
      matched = I18n.available_locales.find { |l| l.to_s.casecmp?(normalized) }
      return matched if matched
    end

    nil
  end
end
