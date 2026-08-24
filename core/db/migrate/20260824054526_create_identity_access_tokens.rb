class CreateIdentityAccessTokens < ActiveRecord::Migration[8.2]
  def change
    create_table :identity_access_tokens, id: :uuid do |t|
      t.references :identity, null: false, foreign_key: true, type: :uuid
      t.string :token, null: false, index: { unique: true }
      t.string :description
      t.string :permission, null: false, default: "write"
      t.datetime :last_used_at

      t.timestamps
    end
  end
end

