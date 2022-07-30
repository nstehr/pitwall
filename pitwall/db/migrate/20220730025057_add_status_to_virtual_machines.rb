class AddStatusToVirtualMachines < ActiveRecord::Migration[7.0]
  def change
    add_column :virtual_machines, :status, :string
  end
end
