class AddScheduledForIndexToBeeperRuns < ActiveRecord::Migration[8.2]
  def change
    add_index :beeper_runs, :scheduled_for
  end
end
