class ZitiApi
   def create_identity(name, role_attributes)
    token = authenticate()

    headers = {'Content-Type' => 'application/json', 'zt-session' => token }

    payload = {
      'enrollment' => {'ott' => true},
       'isAdmin' => false, 
       'name' => name, 
      'type' => 'User',
      'roleAttributes' => role_attributes
    }

    controller = Rails.configuration.x.ziti.controller
    options = { body: payload.to_json,
    headers: headers,
    :verify => false
  }
   identities_url = "#{controller}/edge/management/v1/identities"
   response = HTTParty.post(identities_url, options)
   # {"data":{"_links":{"self":{"href":"./identities/VIntnPjdF0"}},"id":"VIntnPjdF0"},"meta":{}}
   if response.code == 201
      return response["data"]["id"]
   else
      #TODO: handle way better
      puts response
      raise("Error creating identity")
   end
   end

   def get_identity(id)
      token = authenticate()
  
      headers = {'Content-Type' => 'application/json', 'zt-session' => token }
  
      controller = Rails.configuration.x.ziti.controller
      options = { 
      headers: headers,
      :verify => false
    }
     identities_url = "#{controller}/edge/management/v1/identities/#{id}"
     response = HTTParty.get(identities_url, options)

     if response.code == 200
        return response["data"]
     else
        #TODO: handle way better
        puts response
        raise("Error getting identity")
     end
     end

   private
   def authenticate()
    controller = Rails.configuration.x.ziti.controller
    user = Rails.configuration.x.ziti.user
    password = Rails.configuration.x.ziti.pass

    auth_url = "#{controller}/edge/management/v1/authenticate?method=password"

    options = { body: {username: user, password: password}.to_json,
    headers: {'Content-Type' => 'application/json'},
    :verify => false
  }
   results = HTTParty.post(auth_url, options)
   return results.headers["zt-session"]
   end
end