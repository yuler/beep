require_relative "boot"
require "rails/all"

# Require the gems listed in Gemfile, including any gems
# you've limited to :test, :development, or :production.
Bundler.require(*Rails.groups)

# Monorepo env lives at the repo root. Rails.root is `core/`, so the gem's
# default `.env*` lookup would miss `.env` / `.env.local`. First file wins
# unless overwrite is on; overwrite lets `.env.local` replace a stale
# `mise activate` / `.env` value (same overlay as `scripts/dev.sh`).
if defined?(Dotenv::Rails) && !Rails.env.test?
  repo_root = File.expand_path("../..", __dir__)
  Dotenv::Rails.files = [
    "#{repo_root}/.env.#{Rails.env}.local",
    "#{repo_root}/.env.local",
    "#{repo_root}/.env.#{Rails.env}",
    "#{repo_root}/.env"
  ]
  Dotenv::Rails.overwrite = true
end

module BeepCore
  class Application < Rails::Application
    # Initialize configuration defaults for originally generated Rails version.
    config.load_defaults 8.1

    # Please, add to the `ignore` list any other `lib` subdirectories that do
    # not contain `.rb` files, or that should not be reloaded or eager loaded.
    # Common ones are `templates`, `generators`, or `middleware`, for example.
    config.autoload_lib(ignore: %w[assets tasks rails_ext])

    # Configuration for the application, engines, and railties goes here.
    #
    # These settings can be overridden in specific environments using the files
    # in config/environments, which are processed later.
    #
    # config.time_zone = "Central Time (US & Canada)"
    # config.eager_load_paths << Rails.root.join("extras")

    # Use UUID primary keys for all new tables
    config.generators do |g|
      g.orm :active_record, primary_key_type: :uuid
    end

    # Mission dashboard
    config.mission_control.jobs.http_basic_auth_enabled = false
  end
end
