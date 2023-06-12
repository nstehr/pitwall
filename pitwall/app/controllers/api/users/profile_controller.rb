class Api::Users::ProfileController < Api::ApiController
    before_action :authenticate_user!
    def create_zt_identity
        zt_api = ZitiApi.new
        # coupling magic here to know what the naming convetion should be for the role
        id = zt_api.create_identity(current_user.username,["#{current_user.username}-user"])
        identity = zt_api.get_identity(id)
        identity.delete("_links")
        identity.delete("type")
        render json: identity
    end 
end