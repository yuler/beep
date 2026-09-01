# frozen_string_literal: true

# Custom loader for shared flat JSON locales from monorepo root project.inlang.
# Reads settings from project.inlang/settings.json and
# loads messages from project.inlang/messages/*.json.

module JsonLocaleLoader
  def self.inlang_dir
    [
      Rails.root.join("project.inlang"),
      Rails.root.join("../project.inlang")
    ].find { |path| Dir.exist?(path) }
  end

  def self.load_json_locales
    inlang_dir = self.inlang_dir
    return unless inlang_dir

    settings_file = inlang_dir.join("settings.json")
    if File.exist?(settings_file)
      begin
        settings = JSON.parse(File.read(settings_file))
        locales = settings["locales"] || settings["languageTags"]
        base = settings["baseLocale"] || settings["sourceLanguageTag"]
        I18n.available_locales = locales.map(&:to_sym) if locales
        I18n.default_locale = base.to_sym if base
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
        converted = flat_hash.transform_values do |val|
          val.is_a?(String) ? val.gsub(/\{([a-zA-Z0-9_]+)\}/, '%{\1}') : val
        end
        I18n.backend.store_translations(locale, converted.transform_keys(&:to_sym))
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
