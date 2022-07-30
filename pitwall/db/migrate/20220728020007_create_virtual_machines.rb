class CreateVirtualMachines < ActiveRecord::Migration[7.0]
  def change
    create_table :virtual_machines do |t|
      t.string :image
      t.timestamps
      t.belongs_to :orchestrator, index: true, foreign_key: true
    end
  end
end
