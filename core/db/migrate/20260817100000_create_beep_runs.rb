class CreateBeepRuns < ActiveRecord::Migration[8.2]
  def change
    create_table :beep_runs, id: :uuid do |t|
      t.references :beep, null: false, foreign_key: true, type: :uuid, index: false
      t.datetime :scheduled_for, null: false
      t.string :status, null: false, default: "pending"
      t.json :result

      t.timestamps
    end

    add_index :beep_runs, [ :beep_id, :scheduled_for ], unique: true
  end
end
