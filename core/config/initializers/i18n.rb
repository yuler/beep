# frozen_string_literal: true

# Custom loader for shared JSON locales without needing a top-level locale key.
# Maps packages/locales/src/en.json -> I18n translations under :en
# and packages/locales/src/zh-CN.json -> I18n translations under :"zh-CN"

module JsonLocaleLoader
  def self.load_json_locales
    locales_dir = Rails.root.join("../packages/locales/src")
    return unless Dir.exist?(locales_dir)

    I18n.available_locales = %i[en zh-CN]
    I18n.default_locale = :en
    I18n.fallbacks = [:en]

    Dir.glob("#{locales_dir}/*.json").each do |file|
      locale = File.basename(file, ".json").to_sym
      begin
        raw = JSON.parse(File.read(file))
        # Deep symbolize keys
        symbolized = raw.deep_symbolize_keys
        I18n.backend.store_translations(locale, symbolized)
      rescue StandardError => e
        Rails.logger.error("[i18n] Failed to load #{file}: #{e.message}")
      end
    end
  end
end

Rails.application.config.after_initialize do
  JsonLocaleLoader.load_json_locales
end

# Re-load in development on reload
Rails.application.config.to_prepare do
  JsonLocaleLoader.load_json_locales
end
