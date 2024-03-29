package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/nstehr/pitwall/powerunit/core"
)

func main() {
	log.Println("powerunit init system v0.1")
	err := core.SetHostname("foo")
	if err != nil {
		log.Printf("error setting hostname: %s\n", err)
	}
	err = core.MountAll()
	if err != nil {
		log.Printf("Error mounting directories: %s\n", err)
	}
	err = core.LinkAll()
	if err != nil {
		log.Printf("Error linking directories: %s\n", err)
	}
	err = core.SetPermissions()
	if err != nil {
		log.Printf("Error setting permissions on directories: %s\n", err)
	}

	err = core.LinkNameservers()
	if err != nil {
		log.Printf("Error linking nameservers: %s\n", err)
	}

	// drop into the shell for now, just for testing
	cmd := exec.Command("/bin/sh")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// launch an in-process SSH TTY.  There are probably
	// much better ways to do this, but C'est la vie
	go func() {
		err := core.SSH(2222)
		log.Println("Error starting SSH: ", err)
	}()
	// not the way we want to do it properly, but launch the shell and wait
	// ideally we'll want to spawn the processes and have out process here
	// just wait for a signal
	err = cmd.Start()
	if err != nil {
		log.Fatalf("could not start shell: %s", err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("error running shell: %s", err)
	}
}
