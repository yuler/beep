# Hotwire Spark provides dev live-reloading for this importmap-based app.
#
# Two fixes vs. the gem defaults:
#   1. The gem's `configure_jsbundling` runs because the `Jsbundling` constant is
#      present in the bundle, which wrongly points CSS watching at `app/assets/builds`
#      (used by jsbundling/esbuild) instead of `app/assets/stylesheets`. This app
#      serves plain stylesheets, so we restore the importmap-friendly paths.
#   2. Spark's own Action Cable server inherits the app's `solid_cable` adapter,
#      whose isolated server instance fails to broadcast ("No unique index found
#      for id"). Use the in-memory :async adapter for Spark's cable server.

Rails.application.config.to_prepare do
  next unless defined?(Hotwire::Spark) && Hotwire::Spark.enabled?

  Hotwire::Spark.css_paths            = %w[ app/assets/stylesheets ]
  Hotwire::Spark.css_extensions      = %w[ css ]
  Hotwire::Spark.stimulus_paths      = %w[ app/javascript/controllers ]
  Hotwire::Spark.stimulus_extensions = %w[ js ]
  Hotwire::Spark.html_paths          = %w[
    app/controllers app/helpers app/assets/images
    app/models app/views config/locales
  ]
  Hotwire::Spark.html_extensions = %w[ rb erb png jpg jpeg webp svg yaml yml ]
  Hotwire::Spark.html_reload_method = :morph

  # Override the cable server to use async adapter instead of solid_cable.
  # The gem's initialize copies ActionCable::Server::Base.config which inherits
  # solid_cable from cable.yml. We need async for Spark's isolated server.
  Hotwire::Spark::ActionCable::Server.class_eval do
    def initialize(*)
      config = ActionCable::Server::Base.config.dup
      config.connection_class = -> { ActionCable::Connection::Base }
      config.cable = { "adapter" => "async" }
      super(config:)
    end
  end
  Hotwire::Spark.instance_variable_set(:@server, nil)
end
