require 'jwt'
require 'openssl'
require 'net/http'

# based on: https://blog.plataformatec.com.br/2019/01/custom-authentication-methods-with-devise/
class ApiTokenStrategy < Warden::Strategies::Base
    def valid?
      api_token.present?
    end
  
    def authenticate!
      key_string = get_public_key
      # hack the key into PEM format
      pk = OpenSSL::PKey::RSA.new("-----BEGIN PUBLIC KEY-----\n#{key_string}\n-----END PUBLIC KEY-----\n")
      decoded_token = JWT.decode api_token, pk, true, { algorithm: 'RS256' }
      email = decoded_token[0]["email"]
      user = User.find_by(email: email)
  
      if user
        success!(user)
      else
        fail!('Invalid email or password')
      end
    end
  
    private

    def get_public_key
        # this key needs to be fetched from keycloak
        url =  Rails.configuration.x.keycloak.realm_api
        response = HTTParty.get(url)
        response["public_key"] if response.code == 200
    end
  
    def api_token
      env['HTTP_AUTHORIZATION'].to_s.remove('Bearer ')
    end
  end