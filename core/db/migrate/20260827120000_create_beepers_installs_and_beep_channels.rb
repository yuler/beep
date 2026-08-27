class CreateBeepersInstallsAndBeepChannels < ActiveRecord::Migration[8.2]
  def change
    create_table :beepers, id: :uuid do |t|
      t.string :slug, null: false
      t.references :account, foreign_key: true, type: :uuid
      t.string :version, null: false
      t.json :manifest, null: false, default: {}
      t.text :source
      t.timestamps
    end
    add_index :beepers, [ :slug, :account_id ], unique: true

    create_table :beeper_installs, id: :uuid do |t|
      t.references :account, null: false, foreign_key: true, type: :uuid
      t.references :beeper, null: false, foreign_key: true, type: :uuid
      t.string :title, null: false
      t.json :config, default: {}
      t.string :cron, null: false
      t.string :timezone, null: false
      t.string :status, null: false, default: "active"
      t.datetime :next_run_at
      t.datetime :last_run_at
      t.string :alert_state, null: false, default: "ok"
      t.integer :consecutive_failures, null: false, default: 0
      t.integer :schedule_offset, null: false, default: 0
      t.string :ping_token
      t.datetime :last_ping_at
      t.json :notification_channels, null: false, default: []
      t.timestamps
    end
    add_index :beeper_installs, :ping_token, unique: true
    add_index :beeper_installs, [ :account_id, :status, :next_run_at ], name: "index_beeper_installs_on_due"

    create_table :beeper_runs, id: :uuid do |t|
      t.references :beeper_install, null: false, foreign_key: true, type: :uuid
      t.datetime :scheduled_for, null: false
      t.string :status, null: false, default: "pending"
      t.string :check_status
      t.json :check_result
      t.timestamps
    end
    add_index :beeper_runs, [ :beeper_install_id, :scheduled_for ], unique: true

    add_column :beeps, :notification_channels, :json, null: false, default: []
    add_reference :beeps, :beeper_install, foreign_key: true, type: :uuid, null: true
  end
end
