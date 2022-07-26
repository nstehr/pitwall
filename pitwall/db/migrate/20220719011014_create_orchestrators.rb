class CreateOrchestrators < ActiveRecord::Migration[7.0]
  def change
    create_table :orchestrators do |t|
      t.string :status
      t.string :name
    
      t.timestamps
      t.index [:name], unique: true
    end
  end
end
