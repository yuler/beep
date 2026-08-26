json.extract! @account, :id, :name, :slug, :personal
json.notification_channels Current.user.notification_channels
json.timezone Current.user.timezone
json.timezone_source Current.user.timezone_source
