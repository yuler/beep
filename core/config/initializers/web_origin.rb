Rails.application.configure do
  config.x.web_origin = ENV["WEB_URL"].presence || begin
    if Rails.env.production?
      "https://#{ENV.fetch("SITE_DOMAIN")}"
    else
      host = ENV.fetch("APP_HOST", "beep.localhost")
      port = ENV.fetch("WEB_PORT", "3000")
      "http://web.#{host}:#{port}"
    end
  end
end
