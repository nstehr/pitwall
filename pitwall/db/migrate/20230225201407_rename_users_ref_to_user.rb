class RenameUsersRefToUser < ActiveRecord::Migration[7.0]
  def change
    rename_column :virtual_machines, :users_id, :user_id
  end
end
