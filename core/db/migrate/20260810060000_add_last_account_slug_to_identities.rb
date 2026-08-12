class AddLastAccountSlugToIdentities < ActiveRecord::Migration[8.1]
  def change
    add_column :identities, :last_account_slug, :string
  end
end
