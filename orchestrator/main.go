package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/nstehr/pitwall/orchestrator/vm"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	vm.RegisterVMServiceServer(grpcServer, vm.NewApiServer())
	log.Printf("Starting API server on port:%d", *port)
	grpcServer.Serve(lis)
}
