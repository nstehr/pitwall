class ApplicationController < ActionController::Base
    include Pagy::Backend
    def new_session_path(scope)
        new_user_session_path
    end
end
