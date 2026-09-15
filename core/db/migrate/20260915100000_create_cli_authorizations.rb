class CreateCliAuthorizations < ActiveRecord::Migration[8.2]
  def change
    create_table :cli_authorizations, id: :uuid do |t|
      t.references :identity, foreign_key: true, type: :uuid
      t.references :access_token, foreign_key: { to_table: :identity_access_tokens }, type: :uuid
      t.string :device_code, null: false
      t.string :user_code, null: false
      t.string :client_name
      t.string :status, null: false, default: "pending"
      t.datetime :expires_at, null: false
      t.datetime :last_polled_at

      t.timestamps
    end

    add_index :cli_authorizations, :device_code, unique: true
    add_index :cli_authorizations, :user_code, unique: true
    add_index :cli_authorizations, :expires_at
  end
end
