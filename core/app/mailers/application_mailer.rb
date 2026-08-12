class ApplicationMailer < ActionMailer::Base
  default from: email_address_with_name(ENV.fetch("EMAIL_SENDER", "sender@example.com"), ENV.fetch("SITE_NAME", "beep"))

  layout "mailer"
end
