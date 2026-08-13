# Rails session jar (`_beep_session`) — flash, return_to, etc.
# Align Domain with session_id / pending auth for Mode A.
Rails.application.config.session_store :cookie_store, **{
  key: "_#{Rails.application.railtie_name.chomp("_application")}_session",
  same_site: :lax,
  secure: !Rails.env.local?,
  domain: ENV["SESSION_COOKIE_DOMAIN"].presence
}.compact
