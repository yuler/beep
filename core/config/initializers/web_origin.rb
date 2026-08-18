Rails.application.configure do
  config.x.web_origin = ENV["WEB_URL"].presence || begin
    if Rails.env.production?
      "https://#{ENV.fetch("SITE_DOMAIN")}"
    elsif Rails.env.test?
      "http://www.example.com"
    else
      "http://web.#{ENV.fetch("APP_HOST")}:#{ENV.fetch("WEB_PORT")}"
    end
  end
end
