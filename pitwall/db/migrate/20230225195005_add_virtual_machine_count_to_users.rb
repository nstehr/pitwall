class AddVirtualMachineCountToUsers < ActiveRecord::Migration[7.0]
  def change
    add_column :users, :virtual_machines_count, :integer
  end
end
