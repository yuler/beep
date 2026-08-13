class AddMessageToBeeps < ActiveRecord::Migration[8.2]
  def change
    add_column :beeps, :message, :text, null: false
  end
end
