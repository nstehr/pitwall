package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nstehr/pitwall/terminator/stream"
	"github.com/nstehr/pitwall/terminator/vm"
	"github.com/nstehr/pitwall/terminator/ziti"
	zitiSdk "github.com/openziti/sdk-golang/ziti"
	"google.golang.org/protobuf/proto"
)

func main() {

	zitiController, present := os.LookupEnv("ZITI_CONTROLLER")
	if !present {
		log.Fatal("Must specify ZITI_CONTROLLER environment variable")
	}
	zitiUser, present := os.LookupEnv("ZITI_USER")
	if !present {
		log.Fatal("Must specify ZITI_USER environment variable")
	}
	zitiPass, present := os.LookupEnv("ZITI_PASS")
	if !present {
		log.Fatal("Must specify ZITI_PASS environment variable")
	}
	client, err := ziti.NewClient(zitiController, zitiUser, zitiPass)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Login()
	if err != nil {
		log.Fatal(err)
	}
	id, err := client.CreateIdentity(ziti.Device, "nstehr-vm-vmName", false, []string{"nstehr-vm"})
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := client.EnrollIdentity(id)
	if err != nil {
		log.Fatal(err)
	}

	serviceId, err := client.CreateService("nstehr-vm-ssh", true, []string{"nstehr-vm-services"})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(serviceId)

	servicePolicyId, err := client.CreateServicePolicy(ziti.Bind, "nstehr-vm-bind-policy", []string{"#nstehr-vm"}, []string{"#nstehr-vm-services"}, ziti.AllOf)

	if err != nil {
		log.Fatal(err)
	}
	log.Println(servicePolicyId)

	// NOTE TO SELF, should I create the dial policy here too?
	// ./ziti edge create service-policy nstehr-vm-dial-policy Dial --identity-roles '#nstehr-user' --service-roles '#nstehr-vm-services'

	ztx := zitiSdk.NewContextWithConfig(cfg)
	err = ztx.Authenticate()
	if err != nil {
		log.Fatalf("failed to authenticate: %v", err)
	}

	lis, err := ztx.Listen("nstehr-vm-ssh")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := lis.Accept()
			if err != nil {
				log.Println("Error making connection")
			}
			// Handle connections in a new goroutine.
			go handle(conn)
		}
	}()

	stream.RegisterHandler("orchestrator.vm.status.terminator", "orchestrator.vm.status.RUNNING", func(msg []byte) {
		virtualMachine := vm.VM{}
		err := proto.Unmarshal(msg, &virtualMachine)
		if err != nil {
			log.Println("Error unmarshaling VM", err)
			return
		}
		log.Println(virtualMachine.ImageName)
	})
	// Clean exit.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	log.Println("Ctrl-c detected, shutting down")
	log.Println("Goodbye.")

}

func handle(conn net.Conn) {
	vmConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", "172.30.0.2", 2222))
	if err != nil {
		log.Println("error making forwarding connection: ", err)
		return
	}
	connClosed := make(chan struct{}, 1)
	vmConnClosed := make(chan struct{}, 1)
	// the forward function will issue the close
	go forward(conn, vmConn, vmConnClosed)
	go forward(vmConn, conn, connClosed)

	select {
	case <-connClosed:
		log.Println("Incoming connection has closed connection")
	case <-vmConnClosed:
		log.Println("VM connection closed")
	}
}

func forward(from net.Conn, to net.Conn, closedChan chan struct{}) {
	_, err := io.Copy(from, to)

	if err != nil {
		log.Println("Error: ", err)
	}
	if err := from.Close(); err != nil {
		log.Println("Error closing connection: ", err)
	}
	closedChan <- struct{}{}
}
