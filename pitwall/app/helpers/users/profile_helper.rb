require "ssh_data"

module Users::ProfileHelper
    def get_public_key_algo(public_key)
        parsed = SSHData::PublicKey.parse_openssh(public_key)
        return parsed.algo
    end
    def get_public_key_fingerprint(public_key)
        parsed = SSHData::PublicKey.parse_openssh(public_key)
        return parsed.fingerprint
    end
end