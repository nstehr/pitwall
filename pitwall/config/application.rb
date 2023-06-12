require_relative "boot"

require "rails/all"

# Require the gems listed in Gemfile, including any gems
# you've limited to :test, :development, or :production.
Bundler.require(*Rails.groups)

module Pitwall
  class Application < Rails::Application
    # Initialize configuration defaults for originally generated Rails version.
    config.load_defaults 7.0

    # Configuration for the application, engines, and railties goes here.
    #
    # These settings can be overridden in specific environments using the files
    # in config/environments, which are processed later.
    #
    # config.time_zone = "Central Time (US & Canada)"
    # config.eager_load_paths << Rails.root.join("extras")

    # TODO: should this be somewhere else (initializer?)
    config.x.keycloak.client = 'pitwall-ui'
    config.x.keycloak.realm = 'pitwall'
    config.x.keycloak.root_url = ENV['KEYCLOAK_URL']
    config.x.keycloak.realm_api =  "#{config.x.keycloak.root_url}/realms/#{config.x.keycloak.realm }/"

    config.x.rabbitmq.server = ENV['AMQP_SERVER']
    config.x.rabbitmq.username = ENV['AMQP_USER']
    config.x.rabbitmq.password = ENV['AMQP_PASSWORD']
    config.x.rabbitmq.vhost = ENV['AMQP_VHOST']

    config.x.ziti.controller = ENV['ZITI_CONTROLLER']
    config.x.ziti.user = ENV['ZITI_USER']
    config.x.ziti.pass = ENV['ZITI_PASS']
  end
end
