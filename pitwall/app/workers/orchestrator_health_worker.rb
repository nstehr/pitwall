class OrchestratorHealthWorker
    include Sneakers::Worker
  
    from_queue "orchestrator.health", env: nil
    def work(health)
      logger.debug("asfdsdfasdf")
      ack! 
    end
  end