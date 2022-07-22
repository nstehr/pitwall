namespace :rabbitmq do
    desc "Connect consumer to producer"
    task :setup do
      require "bunny"
      exchange_name = "pitwall.orchestration"
      conn = Bunny.new.tap(&:start)
      ch = conn.create_channel
      # create the exchange
      exchange = ch.topic(exchange_name, :durable => true)
      queue_health = ch.queue("orchestrator.health", :durable => true)
      queue_status = ch.queue("orchestrator.vm.status", :durable => true)
      # bind queue to exchange
      queue_health.bind(exchange, routing_key: "orchestrator.health")
      queue_status.bind(exchange, routing_key: "orchestrator.vm.status")
      conn.close
    end
  end