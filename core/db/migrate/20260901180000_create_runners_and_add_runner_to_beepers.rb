class CreateRunnersAndAddRunnerToBeepers < ActiveRecord::Migration[8.2]
  def change
    create_table :runners, id: :uuid do |t|
      t.references :account, null: false, foreign_key: true, type: :uuid
      t.string :name, null: false
      t.string :token_digest, null: false
      t.string :token_prefix, null: false
      t.string :status, null: false, default: "offline"
      t.json :tags, null: false, default: []
      t.boolean :allow_exec, null: false, default: false
      t.string :version
      t.string :os
      t.string :arch
      t.string :hostname
      t.string :ip_address
      t.datetime :last_seen_at

      t.timestamps
    end

    add_index :runners, :token_digest, unique: true
    add_index :runners, [ :account_id, :status ]

    add_reference :beepers, :runner, type: :uuid, foreign_key: true, null: true
    add_column :beepers, :runner_tag, :string, null: true
    add_index :beepers, :runner_tag

    add_reference :beeper_runs, :runner, type: :uuid, foreign_key: true, null: true
    add_column :beeper_runs, :claimed_at, :datetime
  end
end
