class RenameBeepsMessageToTitleAndAddBody < ActiveRecord::Migration[8.2]
  def change
    rename_column :beeps, :message, :title
    add_column :beeps, :body, :text

    reversible do |dir|
      dir.up do
        execute <<~SQL.squish
          UPDATE beeps SET title = substr(title, 1, 80) WHERE length(title) > 80
        SQL
      end
    end
  end
end
