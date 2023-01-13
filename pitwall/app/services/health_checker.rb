class HealthChecker 
    def check(orchestrator)
        begin
            response = Net::HTTP.get_response(URI.parse(orchestrator.healthCheck))
            return true
        # bringing the hammer, just blindly saying all errors are a health check failure
        rescue
            return false
        end
    end

    def schedule(orchestrator)
        rabbit = Rabbitmq.new()
        # we've set up this key to be a delayed 
        message = ::Orch::Orchestrator.encode(orchestrator)
        rabbit.send("orchestrator.healthcheck.schedule", message)
    end
end