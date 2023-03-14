class Users::ProfileController < ApplicationController
    before_action :authenticate_user!
    def show
    end
  
    def show_public_key
        if current_user.identity == nil
            @identity = Identity.new
        else
            @identity = current_user.identity
        end
    end

    def update_public_key
       public_key = params["identity"]["public_key"]
       if current_user.identity == nil
           @identity = Identity.new(user_id: current_user.id, public_key: public_key)
       else
            @identity = current_user.identity
            @identity.public_key = public_key
       end
       @identity.save
        if @identity.valid?
            redirect_to profile_path, notice: "Public Key Updated"
        else
            render :show_public_key, status: :unprocessable_entity
        end
    end
end