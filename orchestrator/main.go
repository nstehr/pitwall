package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"

	"google.golang.org/grpc"

	"github.com/nstehr/pitwall/orchestrator/vm"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

const (
	// executableMask is the mask needed to check whether or not a file's
	// permissions are executable.
	executableMask = 0111

	firecrackerDefaultPath = "firecracker"
)

func main() {
	flag.Parse()

	verifyFirecrackerExists()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	vm.RegisterVMServiceServer(grpcServer, vm.NewApiServer())
	log.Printf("Starting API server on port: %d", *port)
	grpcServer.Serve(lis)
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
