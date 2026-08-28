class BeeperApp < ApplicationRecord
  self.table_name = "beeper_apps"

  belongs_to :account, optional: true
  has_many :beepers, dependent: :destroy

  validates :slug, presence: true, format: { with: /\A[a-z0-9]+(?:-[a-z0-9]+)*\z/, message: "must be lowercase kebab-case" }
  validates :version, presence: true
  validates :slug, uniqueness: { scope: :account_id, message: "has already been taken for this account" }
  validate :validate_manifest_contract

  scope :official, -> { where(account_id: nil) }

  class << self
    def official_dir
      Rails.root.join("../beeper_apps").cleanpath
    end

    def seed_official(logger: Rails.logger)
      return { created: 0, updated: 0, unchanged: 0 } unless Dir.exist?(official_dir)

      stats = { created: 0, updated: 0, unchanged: 0 }

      Dir.glob(official_dir.join("*/manifest.json")).each do |file_path|
        manifest_data = JSON.parse(File.read(file_path))
        slug = manifest_data["slug"]
        version = manifest_data["version"]

        beeper_app = official.find_or_initialize_by(slug: slug)
        is_new = beeper_app.new_record?

        beeper_app.version = version
        beeper_app.manifest = manifest_data

        if is_new
          beeper_app.save!
          stats[:created] += 1
          logger&.info "[BeeperApp] Created official beeper app: #{slug} (#{version})"
        elsif beeper_app.has_changes_to_save?
          beeper_app.save!
          stats[:updated] += 1
          logger&.info "[BeeperApp] Updated official beeper app: #{slug} (#{version})"
        else
          stats[:unchanged] += 1
        end
      end

      stats
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

  def alert_policy_config
    manifest["alerting"] || {}
  end

  def failure_threshold
    manifest.dig("alerting", "failure_threshold") || 2
  end

  def recovery_threshold
    manifest.dig("alerting", "recovery_threshold") || 1
  end

  def min_interval_seconds
    manifest.dig("schedule", "min_interval_seconds") || 60
  end

  def capabilities
    manifest["capabilities"] || []
  end

  def webhook_ping?
    capabilities.include?("webhook_ping")
  end

  def inputs
    manifest["inputs"] || []
  end

  def metrics
    manifest["metrics"] || []
  end

  def receiver_class
    "BeeperApp::Receivers::#{slug.tr("-", "_").camelize}".safe_constantize
  end

  def receiver_available?
    receiver_class.present?
  end

  def produce_signal(config:)
    klass = receiver_class
    if klass
      klass.call(config: config)
    else
      BeeperApp::Signal.new(
        status: :error,
        title: "Beeper app implementation not found",
        message: "No receiver class implemented for beeper app '#{slug}'"
      )
    end
  rescue StandardError => e
    BeeperApp::Signal.new(
      status: :error,
      title: "Signal production failed",
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
