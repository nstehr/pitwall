package core

import (
	"log"
	"os"
	"syscall"
)

func SetHostname(hostname string) error {
	log.Printf("Setting hostname to: %s", hostname)
	err := syscall.Sethostname([]byte(hostname))
	return err
}

func LinkNameservers() error {
	// firecracker will set the provided nameservers to /proc/net/pnp and
	// recommend linking that file to /etc/resolve.conf
	log.Println("Making /etc/resolv.conf a symlink to /proc/net/pnp")
	if _, err := os.Lstat("/etc/resolv.conf"); err == nil {
		// link exists
		log.Println("/etc/resolv.conf exists; removing...")
		err := os.Remove("/etc/resolv.conf")
		if err != nil {
			return err
		}
	}
	err := os.Symlink("/proc/net/pnp", "/etc/resolv.conf")
	return err
}
