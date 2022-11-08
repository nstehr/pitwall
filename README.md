## Pitwall

Yet another system for deploying docker containers as Firecracker VMs

Tested with: Firecracker v1.1.0

# SSH to VM from host
ssh -i id_firecracker -p 2222 -o 'PubkeyAcceptedKeyTypes +ssh-rsa' fred@172.30.0.3
