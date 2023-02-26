class AddNameToVirtualMachines < ActiveRecord::Migration[7.0]
  def change
    add_column :virtual_machines, :name, :string
  end
end
