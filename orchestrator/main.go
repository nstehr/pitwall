package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/nstehr/pitwall/orchestrator/orchestrator"
)

var (
	name = flag.String("name", "", "unique name for the orchestrator, defaults to hostname")
)

const (
	// executableMask is the mask needed to check whether or not a file's
	// permissions are executable.
	executableMask = 0111

	firecrackerDefaultPath = "firecracker"
)

func main() {
	flag.Parse()
	if *name == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Println("could not retrieve hostname")
			*name = "toto"
		}
		*name = hostname
	}

	verifyFirecrackerExists()
	orchestrator.SignalOrchestratorAlive(*name)

	// Clean exit.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	// Exit by user
	log.Println("Ctrl-c detected, shutting down")

	log.Println("Goodbye.")

}

func verifyFirecrackerExists() {
	var firecrackerBinary string

	firecrackerBinary, err := exec.LookPath(firecrackerDefaultPath)
	if err != nil {
		log.Fatalf("failed to lookup firecracker path: %v", err)
	}
	finfo, err := os.Stat(firecrackerBinary)
	if os.IsNotExist(err) {
		log.Fatalf("Binary %q does not exist: %v", firecrackerBinary, err)
	}

	if err != nil {
		log.Fatalf("Failed to stat binary, %q: %v", firecrackerBinary, err)
	}

	if finfo.IsDir() {
		log.Fatalf("Binary, %q, is a directory", firecrackerBinary)
	} else if finfo.Mode()&executableMask == 0 {
		log.Fatalf("Binary, %q, is not executable. Check permissions of binary", firecrackerBinary)
	}
}
