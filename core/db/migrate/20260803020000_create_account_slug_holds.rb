class CreateAccountSlugHolds < ActiveRecord::Migration[8.2]
  def change
    create_table :account_slug_holds, id: :uuid do |t|
      t.references :account, null: false, foreign_key: true, type: :uuid
      t.string :slug, null: false
      t.datetime :expires_at, null: false
      t.timestamps
    end

    add_index :account_slug_holds, :slug
    add_index :account_slug_holds, :expires_at
  end
end
