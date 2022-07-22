class OrchestratorVmStatusWorker
    include Sneakers::Worker
  
    from_queue "orchestrator.vm.status", env: nil
    def work(health)
      #logger.debug("asfdsdfasdf")
      ack! 
    end
  end