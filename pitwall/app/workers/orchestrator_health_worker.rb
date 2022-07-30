require "#{Rails.root}/lib/protos/orchestrator_pb.rb"

class OrchestratorHealthWorker
    include Sneakers::Worker
  
    from_queue "orchestrator.health", env: nil
    def work(health)
      orch = ::Orch::Orchestrator.decode(health)
      Orchestrator.upsert({name: orch.name, status:orch.status}, unique_by: :name)
      ack! 
    end
  end