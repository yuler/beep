class CreateBeeps < ActiveRecord::Migration[8.2]
  def change
    create_table :beeps, id: :uuid do |t|
      t.references :account, null: false, foreign_key: true, type: :uuid
      t.string :kind, null: false # once, recurring
      t.datetime :run_at
      t.datetime :next_run_at
      t.datetime :last_run_at
      t.string :timezone, null: false, default: "UTC"
      t.string :cron
      t.string :status, null: false, default: "active" # active, paused, completed, cancelled

      t.timestamps
    end

    add_index :beeps, [ :status, :next_run_at ]
  end
end
