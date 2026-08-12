class CreateAccountInvitationResponses < ActiveRecord::Migration[8.2]
  def change
    create_table :account_invitation_acceptances, id: :uuid do |t|
      t.references :invitation, null: false, foreign_key: { to_table: :account_invitations }, type: :uuid, index: { unique: true }
      t.references :identity, null: false, foreign_key: true, type: :uuid
      t.references :user, null: false, foreign_key: true, type: :uuid
      t.timestamps
    end

    create_table :account_invitation_declines, id: :uuid do |t|
      t.references :invitation, null: false, foreign_key: { to_table: :account_invitations }, type: :uuid, index: { unique: true }
      t.references :identity, null: false, foreign_key: true, type: :uuid
      t.timestamps
    end

    remove_index :account_invitations, name: "index_account_invitations_on_account_id_and_email"
    add_index :account_invitations, [ :account_id, :email ],
      name: "index_account_invitations_on_account_id_and_email"
  end
end
