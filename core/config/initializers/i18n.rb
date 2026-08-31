# frozen_string_literal: true

# Custom loader for shared flat JSON locales (e.g. "common.save": "Save").
# Expands dot-separated keys into nested hashes so Rails `I18n.t("common.save")`
# works seamlessly.

module JsonLocaleLoader
  def self.load_json_locales
    locales_dir = Rails.root.join("../packages/locales/src")
    return unless Dir.exist?(locales_dir)

    I18n.available_locales = %i[en zh-CN]
    I18n.default_locale = :en
    I18n.fallbacks = [ :en ]

    Dir.glob("#{locales_dir}/*.json").each do |file|
      locale = File.basename(file, ".json").to_sym
      begin
        flat_hash = JSON.parse(File.read(file))
        nested_hash = unflatten_hash(flat_hash)
        I18n.backend.store_translations(locale, nested_hash)
      rescue StandardError => e
        Rails.logger.error("[i18n] Failed to load #{file}: #{e.message}")
      end
    end
  end

  def self.unflatten_hash(flat)
    result = {}
    flat.each do |key, value|
      parts = key.to_s.split(".").map(&:to_sym)
      curr = result
      parts[0...-1].each do |part|
        curr[part] ||= {}
        curr = curr[part]
      end
      curr[parts.last] = value
    end
    result
  end
end

Rails.application.config.after_initialize do
  JsonLocaleLoader.load_json_locales
end

# Re-load in development on reload
Rails.application.config.to_prepare do
  JsonLocaleLoader.load_json_locales
end
