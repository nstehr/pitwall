class Users::OmniauthCallbacksController < Devise::OmniauthCallbacksController
    def keycloakopenid
      @user = User.from_omniauth(request.env["omniauth.auth"])
      if @user.persisted?
        sign_in_and_redirect @user, event: :authentication
      else
        session["devise.keycloakopenid_data"] = request.env["omniauth.auth"]
        redirect_to new_user_registration_url
      end
    end
  
    def failure
      redirect_to root_path
    end
  end