class CreateChannelAuthorizations < ActiveRecord::Migration[8.2]
  def change
    create_table :channel_authorizations, id: :uuid do |t|
      t.references :account, foreign_key: true, type: :uuid
      t.references :user, foreign_key: true, type: :uuid
      t.references :channel, foreign_key: true, type: :uuid
      t.string :device_code, null: false
      t.string :user_code, null: false
      t.string :channel_name
      t.string :status, null: false, default: "pending"
      t.datetime :expires_at, null: false
      t.datetime :last_polled_at

      t.timestamps
    end

    add_index :channel_authorizations, :device_code, unique: true
    add_index :channel_authorizations, :user_code, unique: true
    add_index :channel_authorizations, :expires_at
  end
end
