package core

import (
	"fmt"
	"log"
	"os"
	"syscall"
)

func SetHostname(hostname string) error {
	log.Println(fmt.Sprintf("Setting hostname to: %s", hostname))
	err := syscall.Sethostname([]byte(hostname))
	return err
}

func LinkNameservers() error {
	// firecracker will set the provided nameservers to /proc/net/pnp and
	// recommend linking that file to /etc/resolve.conf
	log.Println("Making /etc/resolv.conf a symlink to /proc/net/pnp")
	err := os.Symlink("/proc/net/pnp", "/etc/resolv.conf")
	return err
}
