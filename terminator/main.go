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
	"github.com/openziti/sdk-golang/ziti/edge"
	"google.golang.org/protobuf/proto"
)

type zitiIds struct {
	vmServiceIds []string
	vmIdentityId string
	lis          edge.Listener
}

// TODO: Refactor to a "ProxiedVM/ZitiedVM struct" with a manager that can handle the
// service creation, tracking, proxying, deleting instead of having it in main.

var vmZitiMap map[int64]zitiIds

func main() {
	vmZitiMap = make(map[int64]zitiIds)
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

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Error getting hostname", err)
	}

	// it's convienient to use the broker to receive each this message, but a bit inefficient, since we
	// are expecting a terminator instance on the same host as the orchestrator.  Could potentially switch
	// to a more local option (zeroMQ, domain socket, etc)
	stream.RegisterHandler(fmt.Sprintf("orchestrator.vm.status.terminator.running.%s", hostname), "orchestrator.vm.status.RUNNING", func(msg []byte) {
		virtualMachine := vm.VM{}
		err := proto.Unmarshal(msg, &virtualMachine)
		if err != nil {
			log.Println("Error unmarshaling VM", err)
			return
		}

		if virtualMachine.Host != hostname {
			return
		}
		client, err := ziti.NewClient(zitiController, zitiUser, zitiPass)
		if err != nil {
			log.Println("Error getting ziti client:", err)
			return
		}
		err = client.Login()
		if err != nil {
			log.Println("Error getting logging into ziti client:", err)
			return
		}
		identityName := fmt.Sprintf("%s-vm-%s", virtualMachine.Owner, virtualMachine.Name)
		identityRole := fmt.Sprintf("%s-vm", virtualMachine.Owner)
		id, err := client.CreateIdentity(ziti.Device, identityName, false, []string{identityRole})
		if err != nil {
			log.Println("Error creating identity:", err)
			return
		}
		cfg, err := client.EnrollIdentity(id)
		if err != nil {
			log.Println("Error enrolling identity:", err)
			return
		}
		// TODO: hardcode ssh as a service, next step would be to iterate from virtualMachine.Services
		serviceName := fmt.Sprintf("%s-vm-%s-ssh", virtualMachine.Owner, virtualMachine.Name)
		serviceRole := fmt.Sprintf("%s-vm-services", virtualMachine.Owner)
		sshServiceId, err := client.CreateService(serviceName, true, []string{serviceRole})
		if err != nil {
			log.Println("Error creating service:", err)
			return
		}

		bindPolicyName := fmt.Sprintf("%s-vm-bind-policy", virtualMachine.Owner)
		identityRoleAttr := fmt.Sprintf("#%s", identityRole)
		serviceRoleAttr := fmt.Sprintf("#%s", serviceRole)
		_, err = client.CreateServicePolicy(ziti.Bind, bindPolicyName, []string{identityRoleAttr}, []string{serviceRoleAttr}, ziti.AllOf)

		if err != nil {
			log.Println("Error creating service policy:", err)
		}

		// NOTE TO SELF, should I create the dial policy here too?
		dialPolicyName := fmt.Sprintf("%s-vm-dial-policy", virtualMachine.Owner)
		userIdentityRoleAttr := fmt.Sprintf("#%s-user", virtualMachine.Owner)
		_, err = client.CreateServicePolicy(ziti.Dial, dialPolicyName, []string{userIdentityRoleAttr}, []string{serviceRoleAttr}, ziti.AllOf)
		if err != nil {
			log.Println("Error creating service policy:", err)
		}

		// we are expecting the config to be in memory, not very crash safe, TODO: persist (in vault?)
		ztx := zitiSdk.NewContextWithConfig(cfg)
		err = ztx.Authenticate()
		if err != nil {
			log.Println("failed to authenticate: ", err)
		}

		// TODO: same as above, we'll iterate the virtualMachine.Services
		lis, err := ztx.Listen(serviceName)
		if err != nil {
			log.Println("failed to listen: ", err)
			return
		}
		// track...TODO: track better?
		vmZitiMap[virtualMachine.Id] = zitiIds{vmIdentityId: id, vmServiceIds: []string{sshServiceId}, lis: lis}

		go func() {
			for {
				// Listen for an incoming connection.
				conn, err := lis.Accept()
				if err != nil {
					log.Println("Error making connection")
					return
				}
				// Handle connections in a new goroutine.
				go handle(conn, virtualMachine.PrivateIp, 2222)
			}
		}()
	})

	stream.RegisterHandler(fmt.Sprintf("orchestrator.vm.status.terminator.stopped.%s", hostname), "orchestrator.vm.status.STOPPED", func(msg []byte) {
		virtualMachine := vm.VM{}
		err := proto.Unmarshal(msg, &virtualMachine)
		if err != nil {
			log.Println("Error unmarshaling VM", err)
			return
		}
		ids, ok := vmZitiMap[virtualMachine.Id]
		if !ok {
			log.Printf("Not tracking virtual machine: %d\n", virtualMachine.Id)
		}
		client, err := ziti.NewClient(zitiController, zitiUser, zitiPass)
		if err != nil {
			log.Println("Error getting ziti client:", err)
			return
		}
		err = client.Login()
		if err != nil {
			log.Println("Error getting logging into ziti client:", err)
			return
		}
		for _, serviceId := range ids.vmServiceIds {
			err := client.DeleteService(serviceId)
			if err != nil {
				log.Println("Error deleting service: ", err)
			}
		}
		// this should be pushed into the loop
		lis := ids.lis
		if lis != nil {
			lis.Close()
		}
	})

	// Clean exit.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	log.Println("Ctrl-c detected, shutting down")
	log.Println("Goodbye.")

}

func handle(conn net.Conn, remoteIp string, remotePort int) {
	vmConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", remoteIp, remotePort))
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
