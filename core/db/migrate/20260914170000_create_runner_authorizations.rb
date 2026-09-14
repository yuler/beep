class CreateRunnerAuthorizations < ActiveRecord::Migration[8.2]
  def change
    create_table :runner_authorizations, id: :uuid do |t|
      t.references :account, foreign_key: true, type: :uuid
      t.references :user, foreign_key: true, type: :uuid
      t.references :runner, foreign_key: true, type: :uuid
      t.string :device_code, null: false
      t.string :user_code, null: false
      t.string :runner_name
      t.json :tags, default: [], null: false
      t.json :metadata, default: {}, null: false
      t.string :status, null: false, default: "pending"
      t.datetime :expires_at, null: false
      t.datetime :last_polled_at

      t.timestamps
    end

    add_index :runner_authorizations, :device_code, unique: true
    add_index :runner_authorizations, :user_code, unique: true
    add_index :runner_authorizations, :expires_at
  end
end
