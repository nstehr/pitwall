require 'sneakers'
require 'bunny'

opts = {
    host: Rails.configuration.x.rabbitmq.server,
    vhost: Rails.configuration.x.rabbitmq.vhost,
    user: Rails.configuration.x.rabbitmq.username,
    password: Rails.configuration.x.rabbitmq.password
}

conn = Bunny.new(opts)
Sneakers.configure(:connection => conn)