class AddVirtualMachineToServices < ActiveRecord::Migration[7.0]
  def change
    add_reference :services, :virtual_machine, null: false, foreign_key: true
  end
end
