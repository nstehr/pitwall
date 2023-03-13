class CreateIdentities < ActiveRecord::Migration[7.0]
  def change
    create_table :identities do |t|
      t.text :public_key
      t.text :ziti_identity
      t.references :user, null: false, foreign_key: true

      t.timestamps
    end
  end
end
