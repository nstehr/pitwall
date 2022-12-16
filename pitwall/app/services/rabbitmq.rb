require 'bunny'

class Rabbitmq
    @@exchange = "pitwall.orchestration"

    def send(routing_key, message)

        opts = {
            host: Rails.configuration.x.rabbitmq.server,
            vhost: Rails.configuration.x.rabbitmq.vhost,
            user: Rails.configuration.x.rabbitmq.username,
            password: Rails.configuration.x.rabbitmq.password
        }

        conn = Bunny.new(opts).tap(&:start)
        ch = conn.create_channel
        exchange = ch.topic(@@exchange, :durable => true)
        exchange.publish(message, routing_key: routing_key)
    end
end