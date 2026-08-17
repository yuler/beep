class CreatePushSubscriptions < ActiveRecord::Migration[8.2]
  def change
    create_table :push_subscriptions, id: :uuid do |t|
      t.references :account, null: false, foreign_key: true, type: :uuid
      t.references :user, null: false, foreign_key: true, type: :uuid, index: false
      t.text :endpoint
      t.string :p256dh_key
      t.string :auth_key
      t.string :user_agent, limit: 4096

      t.timestamps
    end

    add_index :push_subscriptions, [ :user_id, :endpoint ], unique: true
  end
end
