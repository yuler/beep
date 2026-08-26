class CreatePlugins < ActiveRecord::Migration[8.2]
  def change
    create_table :plugins, id: :uuid do |t|
      t.string :slug, null: false
      t.references :account, foreign_key: true, type: :uuid
      t.string :version, null: false
      t.json :manifest, null: false, default: {}
      t.text :source

      t.timestamps
    end

    # slug is unique for official plugins (account_id is null)
    # and unique per account for custom plugins
    add_index :plugins, [ :slug, :account_id ], unique: true
  end
end
