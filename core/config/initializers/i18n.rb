# frozen_string_literal: true

# Custom loader for shared flat JSON locales from monorepo root project.inlang.
# Reads settings from project.inlang/settings.json and
# loads messages from project.inlang/messages/*.json.

module JsonLocaleLoader
  def self.load_json_locales
    inlang_dir = Rails.root.join("../project.inlang")
    return unless Dir.exist?(inlang_dir)

    settings_file = inlang_dir.join("settings.json")
    if File.exist?(settings_file)
      begin
        settings = JSON.parse(File.read(settings_file))
        I18n.available_locales = settings["languageTags"].map(&:to_sym) if settings["languageTags"]
        I18n.default_locale = settings["sourceLanguageTag"].to_sym if settings["sourceLanguageTag"]
        I18n.fallbacks = [ I18n.default_locale ]
      rescue StandardError => e
        Rails.logger.error("[i18n] Failed to parse #{settings_file}: #{e.message}")
      end
    end

    messages_dir = inlang_dir.join("messages")
    return unless Dir.exist?(messages_dir)

    Dir.glob("#{messages_dir}/*.json").each do |file|
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
