require "#{Rails.root}/lib/protos/orchestrator_pb.rb"

class OrchestratorHealthWorker
    include Sneakers::Worker
  
    from_queue "orchestrator.health", env: nil
    def work(health)
      orch = ::Orch::Orchestrator.decode(health)
      data = {name: orch.name, status:orch.status, health_check_url: (orch.healthCheck if !orch.healthCheck.blank?)}.compact
      Orchestrator.upsert(data, unique_by: :name)
      if !orch.healthCheck.blank?
        health_checker = HealthChecker.new
        health_checker.schedule(orch)
      end
      ack! 
    end
  end