# Rails session jar (`_mono_solo_session`) — flash, return_to, etc.
# Align Domain with session_id / pending auth for Mode A.
Rails.application.config.session_store :cookie_store, **{
  key: "_mono_solo_session",
  same_site: :lax,
  secure: !Rails.env.local?,
  domain: ENV["SESSION_COOKIE_DOMAIN"].presence
}.compact
