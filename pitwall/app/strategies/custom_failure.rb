require 'json'
class CustomFailure < Devise::FailureApp
    include ActionController::Head
    include ActionController::MimeResponds

    def respond
        # messing around, trying to prevent devise from returning
        # html error/login page when using the JSON API
        unless request.format.to_sym == :html
            self.status = 401 
            self.content_type = 'json'
            self.response_body = {"errors" => ["Invalid login credentials"]}.to_json
        else
          super
        end
      end
end