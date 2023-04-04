package core

import (
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

func SetPermissions() error {
	// I noticed in some images (like ubuntu) the permissions on some directories weren't
	// right for general VM use, so this should update them
	log.Println("Setting permissions")
	return os.Chmod("/tmp", 01777)
}

func mount(target string, fsType string, flags uintptr) error {

	if _, err := os.Stat(target); os.IsNotExist(err) {
		err := os.MkdirAll(target, 0755)
		if err != nil {
			return err
		}
	}

	log.Printf("Mounting: %s as filesystem type: %s\n", target, fsType)
	err := syscall.Mount("none", target, fsType, flags, "")
	if err != nil {
		return err
	}

	return nil
}

func LinkAll() error {

	// https://utcc.utoronto.ca/~cks/space/blog/unix/DevFdImplementations
	os.Symlink("/proc/self/fd", "/dev/fd")
	os.Symlink("/proc/self/fd/0", "/dev/stdin")
	os.Symlink("/proc/self/fd/1", "/dev/stdout")
	os.Symlink("/proc/self/fd/2", "/dev/stderr")

	return nil
}
