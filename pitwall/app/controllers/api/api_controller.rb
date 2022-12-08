class Api::ApiController < ActionController::API
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
    after_action lambda {
       # HACK! to remove cookie on API requests
       # TODO: this still doesn't work when there is an error, need to look further
       request.session_options[:skip] = true
    }
end