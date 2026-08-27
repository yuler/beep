class Beeper < ApplicationRecord
  belongs_to :account, optional: true

  validates :slug, presence: true, format: { with: /\A[a-z0-9]+(?:-[a-z0-9]+)*\z/, message: "must be lowercase kebab-case" }
  validates :version, presence: true
  validates :slug, uniqueness: { scope: :account_id, message: "has already been taken for this account" }
  validate :validate_manifest_contract

  scope :official, -> { where(account_id: nil) }

  class << self
    def official_dir
      Rails.root.join("../apps/beepers").cleanpath
    end

    def seed_official
      return unless Dir.exist?(official_dir)

      Dir.glob(official_dir.join("*/manifest.json")).each do |file_path|
        manifest_data = JSON.parse(File.read(file_path))
        beeper = official.find_or_initialize_by(slug: manifest_data["slug"])
        beeper.version = manifest_data["version"]
        beeper.manifest = manifest_data
        beeper.save!
      end
    end
  end

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
    "Beeper::Checkers::#{slug.tr("-", "_").camelize}".safe_constantize
  end

  def run_check(config:)
    klass = checker_class
    if klass
      klass.call(config: config)
    else
      Beeper::CheckResult.new(
        status: :error,
        title: "Beeper implementation not found",
        message: "No checker class implemented for beeper '#{slug}'"
      )
    end
  rescue StandardError => e
    Beeper::CheckResult.new(
      status: :error,
      title: "Check execution failed",
      message: e.message
    )
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
