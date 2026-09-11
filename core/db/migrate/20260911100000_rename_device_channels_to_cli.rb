class RenameDeviceChannelsToCli < ActiveRecord::Migration[8.2]
  def up
    execute "UPDATE channels SET kind = 'cli' WHERE kind = 'device'"
  end

  def down
    execute "UPDATE channels SET kind = 'device' WHERE kind = 'cli'"
  end
end
