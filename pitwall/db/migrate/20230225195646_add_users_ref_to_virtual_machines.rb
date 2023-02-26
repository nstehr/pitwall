class AddUsersRefToVirtualMachines < ActiveRecord::Migration[7.0]
  def change
    add_reference :virtual_machines, :users, null: false, foreign_key: true
  end
end
