class Plugin < ApplicationRecord
  self.table_name = "plugins"

  belongs_to :account, optional: true

  scope :official, -> { where(account_id: nil) }
  scope :custom, -> { where.not(account_id: nil) }

  validates :slug, presence: true, format: { with: /\A[a-z0-9]+(?:-[a-z0-9]+)*\z/, message: "must be lowercase kebab-case" }
  validates :version, presence: true
  validates :slug, uniqueness: { scope: :account_id, message: "has already been taken for this account" }
  validate :validate_manifest_contract

  def official?
    account_id.nil?
  end

  def custom?
    !official?
  end

  def name
    manifest["name"] || slug
  end

  def description
    manifest["description"]
  end

  def default_cron
    manifest.dig("schedule", "default_cron")
  end

  def failure_threshold
    manifest.dig("schedule", "failure_threshold") || 2
  end

  def min_interval_seconds
    manifest.dig("schedule", "min_interval_seconds") || 60
  end

  def webhook_ingest?
    manifest.dig("ingest", "webhook") == true
  end

  def inputs
    manifest["inputs"] || []
  end

  def metrics
    manifest["metrics"] || []
  end

  def checker_class
    checker_name = slug.tr("-", "_").camelize
    "Plugin::Checkers::#{checker_name}".safe_constantize
  end

  def run_check(config:)
    klass = checker_class
    if klass
      klass.call(config: config)
    else
      Plugin::CheckResult.new(
        status: :error,
        title: "Plugin implementation not found",
        message: "No checker class implemented for plugin '#{slug}'"
      )
    end
  rescue StandardError => e
    Plugin::CheckResult.new(
      status: :error,
      title: "Check execution failed",
      message: e.message
    )
  end

  class << self
    def official_plugins_dir
      Rails.root.join("../apps/plugins").cleanpath
    end

    def seed_official_plugins!
      return unless Dir.exist?(official_plugins_dir)

      manifest_files = Dir.glob(official_plugins_dir.join("*/manifest.json"))
      manifest_files.each do |file_path|
        raw_json = File.read(file_path)
        manifest_data = JSON.parse(raw_json)
        slug = manifest_data["slug"]
        version = manifest_data["version"]

        plugin = official.find_or_initialize_by(slug: slug)
        plugin.version = version
        plugin.manifest = manifest_data
        plugin.save!
      end
    end
  end

  private

  def validate_manifest_contract
    return if manifest.blank?

    validator = ManifestValidator.new(manifest)
    unless validator.valid?
      validator.errors.each do |err|
        errors.add(:manifest, err)
      end
    end

    if manifest.is_a?(Hash)
      if slug.present? && manifest["slug"].present? && slug != manifest["slug"]
        errors.add(:slug, "must match manifest slug '#{manifest['slug']}'")
      end
      if version.present? && manifest["version"].present? && version != manifest["version"]
        errors.add(:version, "must match manifest version '#{manifest['version']}'")
      end
    end
  end
end
