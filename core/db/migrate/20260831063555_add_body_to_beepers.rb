class AddBodyToBeepers < ActiveRecord::Migration[8.2]
  def change
    add_column :beepers, :body, :text
  end
end
