Rails.application.config.app_version = ENV.fetch("APP_VERSION", "-")
Rails.application.config.git_revision = ENV.fetch("GIT_REVISION", "-")
