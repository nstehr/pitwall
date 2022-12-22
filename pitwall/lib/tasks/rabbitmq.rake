namespace :rabbitmq do
    desc "Connect consumer to producer"
    task :setup do
      require "bunny"
      exchange_name = "pitwall.orchestration"
      opts = {
        host: Rails.configuration.x.rabbitmq.server,
        vhost: Rails.configuration.x.rabbitmq.vhost,
        user: Rails.configuration.x.rabbitmq.username,
        password: Rails.configuration.x.rabbitmq.password
    }
    
      conn = Bunny.new(opts).tap(&:start)
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