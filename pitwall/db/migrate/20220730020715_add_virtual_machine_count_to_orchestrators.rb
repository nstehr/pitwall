class AddVirtualMachineCountToOrchestrators < ActiveRecord::Migration[7.0]
  def change
    add_column :orchestrators, :virtual_machines_count, :integer
  end
end
