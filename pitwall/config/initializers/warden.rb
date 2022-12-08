ActiveSupport::Reloader.to_prepare do
    Warden::Strategies.add(:api_token, ApiTokenStrategy)
  end

