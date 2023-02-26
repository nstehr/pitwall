class AddUniqueNameIndexToVirtualMachines < ActiveRecord::Migration[7.0]
  def change
    add_index :virtual_machines, [:name, :user_id], unique: true
  end
end
