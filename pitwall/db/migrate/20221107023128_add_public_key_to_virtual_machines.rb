class AddPublicKeyToVirtualMachines < ActiveRecord::Migration[7.0]
  def change
    add_column :virtual_machines, :public_key, :string
  end
end
