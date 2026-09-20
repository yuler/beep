class MigrateRunnerIdleStatusToOffline < ActiveRecord::Migration[8.2]
  def up
    execute "UPDATE runners SET status = 'offline' WHERE status = 'idle'"
  end

  def down
  end
end
