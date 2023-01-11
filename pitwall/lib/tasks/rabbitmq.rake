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
      queue_delay =  ch.queue("orchestrator.healthcheck.schedule", arguments: {
        # set the dead-letter exchange to the default queue
        'x-dead-letter-exchange' => exchange_name,
        # when the message expires, set change the routing key into the destination queue name
        'x-dead-letter-routing-key' => 'orchestrator.healthcheck.execute',
        # the time in milliseconds to keep the message in the queue
        'x-message-ttl' => 5000
      })
      queue_work = ch.queue("orchestrator.healthcheck.execute", :durable => true)
      # bind queue to exchange
      queue_health.bind(exchange, routing_key: "orchestrator.health")
      queue_status.bind(exchange, routing_key: "orchestrator.vm.status")
      queue_delay.bind(exchange,  routing_key: "orchestrator.healthcheck.schedule")
      queue_work.bind(exchange,  routing_key: "orchestrator.healthcheck.execute")
      conn.close
    end
  end