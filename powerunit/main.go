package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/nstehr/pitwall/powerunit/core"
)

func main() {
	log.Println("powerunit init system v0.1")
	err := core.SetHostname("foo")
	if err != nil {
		log.Println(fmt.Sprintf("error setting hostname: ", err))
	}
	err = core.MountAll()
	if err != nil {
		log.Println(fmt.Sprintf("Error mounting directories: ", err))
	}

	// drop into the shell for now, just for testing
	cmd := exec.Command("/bin/sh")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// not the way we want to do it properly, but launch the shell and wait
	// ideally we'll want to spawn the processes and have out process here
	// just wait for a signal
	err = cmd.Start()
	if err != nil {
		panic(fmt.Sprintf("could not start shell: %s", err))
	}

	err = cmd.Wait()
	if err != nil {
		panic(fmt.Sprintf("error running shell: %s", err))
	}
}
