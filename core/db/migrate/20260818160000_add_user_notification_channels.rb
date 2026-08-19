class AddUserNotificationChannels < ActiveRecord::Migration[8.2]
  def change
    add_column :users, :notification_channels, :json, null: false, default: [ "email", "web_push" ]
  end
end
