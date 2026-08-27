class CreateBeeperAppsBeepersAndBeepChannels < ActiveRecord::Migration[8.2]
  def change
    create_table :beeper_apps, id: :uuid do |t|
      t.string :slug, null: false
      t.references :account, foreign_key: true, type: :uuid
      t.string :version, null: false
      t.json :manifest, null: false, default: {}
      t.text :source
      t.timestamps
    end
    add_index :beeper_apps, [ :slug, :account_id ], unique: true

    create_table :beepers, id: :uuid do |t|
      t.references :account, null: false, foreign_key: true, type: :uuid
      t.references :beeper_app, null: false, foreign_key: true, type: :uuid
      t.string :title, null: false
      t.json :config, default: {}
      t.string :cron, null: false
      t.string :timezone, null: false
      t.string :status, null: false, default: "active"
      t.datetime :next_run_at
      t.datetime :last_run_at
      t.string :alert_state, null: false, default: "ok"
      t.integer :consecutive_failures, null: false, default: 0
      t.json :signal_metadata, null: false, default: {}
      t.json :notification_channels, null: false, default: []
      t.timestamps
    end
    add_index :beepers, "(json_extract(signal_metadata, '$.ping_token'))", unique: true, name: "index_beepers_on_ping_token"
    add_index :beepers, [ :account_id, :status, :next_run_at ], name: "index_beepers_on_due"

    create_table :beeper_runs, id: :uuid do |t|
      t.references :beeper, null: false, foreign_key: true, type: :uuid
      t.datetime :scheduled_for, null: false
      t.string :status, null: false, default: "pending"
      t.string :signal_status
      t.json :signal_result
      t.timestamps
    end
    add_index :beeper_runs, [ :beeper_id, :scheduled_for ], unique: true

    add_column :beeps, :notification_channels, :json, null: false, default: []
    add_reference :beeps, :beeper, foreign_key: true, type: :uuid, null: true
  end
end
