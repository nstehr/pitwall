package core

import (
	"fmt"
	"log"
	"syscall"
)

func SetHostname(hostname string) error {
	log.Println(fmt.Sprintf("Setting hostname to: %s", hostname))
	err := syscall.Sethostname([]byte(hostname))
	return err
}
