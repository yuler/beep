class AddPluginFieldsToBeepsAndBeepRuns < ActiveRecord::Migration[8.2]
  def change
    add_reference :beeps, :plugin, foreign_key: true, type: :uuid, null: true
    add_column :beeps, :plugin_config, :json
    add_column :beeps, :alert_state, :string, null: false, default: "ok"
    add_column :beeps, :consecutive_failures, :integer, null: false, default: 0
    add_column :beeps, :schedule_offset, :integer, null: false, default: 0
    add_column :beeps, :ping_token, :string
    add_column :beeps, :last_ping_at, :datetime

    add_index :beeps, :ping_token, unique: true

    add_column :beep_runs, :check_status, :string
    add_column :beep_runs, :check_result, :json
  end
end
