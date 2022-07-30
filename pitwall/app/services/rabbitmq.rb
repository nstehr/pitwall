require 'bunny'

class Rabbitmq
    @@exchange = "pitwall.orchestration"

    def self.exchange
        @@exchange
    end

    def send(routing_key, message)
        conn = Bunny.new.tap(&:start)
        ch = conn.create_channel
        exchange = ch.topic(@@exchange, :durable => true)
        exchange.publish(message, routing_key: routing_key)
    end
end