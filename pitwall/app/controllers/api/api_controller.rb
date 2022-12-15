class Api::ApiController < ActionController::API

    around_action :skip_session

    rescue_from StandardError do |exception|
        render json: { :error => exception.message }, :status => 500
    end    
    # from: https://robl.me/posts/the-magical-devise-journey
    # allows for the API to use just the api_token warden
    def authenticate_user!(opts = {})
        opts[:scope] = :user
        warden.authenticate!(:api_token, opts) if !devise_controller? || opts.delete(:force)
    end
    def current_user
        @current_user ||= warden.authenticate(:api_token, scope: :user)
    end

    # kind of hacky, but make sure no cookie/session is used for API requests
    def skip_session
        yield
      ensure
        request.session_options[:skip] = true
      end
end