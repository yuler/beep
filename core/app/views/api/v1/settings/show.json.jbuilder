json.extract! @account, :id, :name, :slug, :personal
json.notification_channels Current.user.notification_channels
