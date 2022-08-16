package core

import (
	"fmt"
	"log"
	"os"
	"syscall"
)

func MountAll() error {
	// From what I gathered, mounting what seem to be neccessary to get going
	mount("/proc", "proc", 0)
	mount("/dev/pts", "devpts", 0)
	mount("/dev/mqueue", "mqueue", 0)
	mount("/dev/shm", "tmpfs", 0)
	mount("/run", "tmpfs", 0)
	mount("/sys", "sysfs", 0)
	mount("/sys/fs/cgroup", "cgroup", 0)

	return nil
}

func mount(target string, fsType string, flags uintptr) error {

	if _, err := os.Stat(target); os.IsNotExist(err) {
		err := os.MkdirAll(target, 0755)
		if err != nil {
			return err
		}
	}

	log.Println(fmt.Sprintf("Mounting: %s as filesystem type: %s", target, fsType))
	err := syscall.Mount("none", target, fsType, flags, "")
	if err != nil {
		return err
	}

	return nil
}
