class AddBeepChannelsAndEmailChannelEnabled < ActiveRecord::Migration[8.2]
  def up
    add_column :accounts, :email_channel_enabled, :boolean, null: false, default: true
    add_column :beeps, :channels, :json

    Beep.reset_column_information
    Beep.includes(:account).find_each do |beep|
      channels = beep.account.personal? ? %w[ email web_push ] : %w[ web_push ]
      beep.update_columns(channels: channels)
    end

    change_column_null :beeps, :channels, false
  end

  def down
    remove_column :beeps, :channels
    remove_column :accounts, :email_channel_enabled
  end
end
