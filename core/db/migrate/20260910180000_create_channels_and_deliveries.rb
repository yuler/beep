class CreateChannelsAndDeliveries < ActiveRecord::Migration[8.2]
  def change
    create_table :channels, id: :uuid do |t|
      t.references :account, null: false, foreign_key: true, type: :uuid
      t.references :user, null: false, foreign_key: true, type: :uuid
      t.string :kind, null: false
      t.string :name, null: false
      t.string :token, null: false
      t.string :status, null: false, default: "active"
      t.json :config, null: false, default: {}
      t.datetime :last_seen_at

      t.timestamps
    end

    add_index :channels, :token, unique: true
    add_index :channels, [ :user_id, :kind ]
    add_index :channels, [ :account_id, :status ]

    create_table :channel_deliveries, id: :uuid do |t|
      t.references :channel, null: false, foreign_key: true, type: :uuid
      t.references :beep_run, foreign_key: true, type: :uuid
      t.string :status, null: false, default: "pending"
      t.json :payload, null: false, default: {}
      t.datetime :claimed_at
      t.datetime :expires_at

      t.timestamps
    end

    add_index :channel_deliveries, [ :channel_id, :status ]
    add_index :channel_deliveries, :expires_at
  end
end
