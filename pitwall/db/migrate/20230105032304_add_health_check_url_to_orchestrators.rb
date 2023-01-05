class AddHealthCheckUrlToOrchestrators < ActiveRecord::Migration[7.0]
  def change
    add_column :orchestrators, :health_check_url, :string
  end
end
