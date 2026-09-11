class AddSourceAndIntentToBeeps < ActiveRecord::Migration[8.2]
  def change
    add_reference :beeps, :source, polymorphic: true, type: :uuid, null: true, index: true
    add_column :beeps, :intent, :string
    add_column :beeps, :metadata, :json, default: {}, null: false

    reversible do |dir|
      dir.up do
        execute <<~SQL
          UPDATE beeps
          SET source_type = 'Beeper', source_id = beeper_id
          WHERE beeper_id IS NOT NULL
        SQL
      end
    end
  end
end
