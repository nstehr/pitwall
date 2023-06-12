class CreateServices < ActiveRecord::Migration[7.0]
  def change
    create_table :services do |t|
      t.text :name
      t.integer :port
      t.boolean :private
      t.string :protocol

      t.timestamps
    end
  end
end
