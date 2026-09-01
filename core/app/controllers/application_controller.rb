class ApplicationController < ActionController::Base
  include Authentication
  include Authorization
  include CurrentRequest # TODO: others: CurrentTimezone, SetPlatform
  include SetLocale
  include RequestForgeryProtection

  include AccountUrlHelper
  # Only allow modern browsers supporting webp images, web push, badges, import maps, CSS nesting, and CSS :has.
  allow_browser versions: :modern
end
