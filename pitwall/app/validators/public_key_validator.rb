require "ssh_data"

class PublicKeyValidator < ActiveModel::EachValidator
    def validate_each(record, attribute, value)
        # allow the value to be optional, and only validate if present. TODO: evaluate if there is a better way
        unless value.blank?
            begin
                SSHData::PublicKey.parse_openssh(value)
            rescue
                record.errors.add attribute, (options[:message] || "invalid public key format")
            end
        end
    end
end