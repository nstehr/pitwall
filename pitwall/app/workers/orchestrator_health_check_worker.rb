require "#{Rails.root}/lib/protos/orchestrator_pb.rb"

class OrchestratorHealthCheckWorker
    include Sneakers::Worker
  
    from_queue "orchestrator.healthcheck.execute", env: nil
    def work(health)
      orch = ::Orch::Orchestrator.decode(health)
      if !orch.healthCheck.blank?
        health_checker = HealthChecker.new
        healthy = health_checker.check(orch)
        if healthy 
            health_checker.schedule(orch)
        else
            puts "health check failed"
        end
      end
      ack! 
    end
  end