package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/coreos/go-iptables/iptables"
	"github.com/nstehr/pitwall/orchestrator/orchestrator"
	"github.com/nstehr/pitwall/orchestrator/vm"
	"github.com/vishvananda/netlink"
)

var (
	name          = flag.String("name", "", "unique name for the orchestrator, defaults to hostname")
	subnet        = flag.String("subnet", "172.30.0.0/16", "subnet to use on host for VM ips")
	hostInterface = flag.String("hostInterface", "wlp2s0", "host interface to use for outbound traffic")
	setup         = flag.Bool("setup", false, "setups up host networking (bridge and iptables rules)")
	healthPort    = flag.Int("healthPort", 9119, "port for health check endpoint")
)

const (
	// executableMask is the mask needed to check whether or not a file's
	// permissions are executable.
	executableMask         = 0111
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

	ipam, err := vm.NewIpam(*subnet)
	if *setup {
		log.Println("Initializing host networking setup")
		if err != nil {
			log.Fatalf("failed to create ipam service: %v", err)
		}
		err = initHostNetworking(ipam, *hostInterface)
		if err != nil {
			log.Fatalf("failed to initialize host networking: %v", err)
		}
		return
	}

	bridge, err := netlink.LinkByName(vm.BridgeName)
	bridgeIf, err := netlink.AddrList(bridge, netlink.FAMILY_V4)
	if err != nil {
		log.Fatalf("error retrieving host bridge IP", err)
	}
	bridgeIfIp := bridgeIf[0].IP
	// TODO: probably a better way, but take the ip acting as the 'gateway' and reserve premptively
	err = ipam.AcquireSpecificIP(bridgeIfIp.String())
	if err != nil {
		log.Fatalf("failed to reserve host gateway: %v", err)
	}
	verifyFirecrackerExists()
	initHealthCheck(*healthPort)
	// bit of a hack to use the host interface for dual purpose, TODO: fix better later :)
	healthUrl := ""
	hostIp, err := getHostIP(*hostInterface)
	if err != nil {
		log.Println(err)
	} else {
		healthUrl = fmt.Sprintf("http://%s:%d/health", hostIp, *healthPort)
	}
	fmt.Printf("Registering orchestrator with name: %s and health check url: %s\n", *name, healthUrl)
	ctx := context.Background()
	err = orchestrator.SignalOrchestratorAlive(ctx, *name, healthUrl)
	if err != nil {
		log.Println(err)
	}

	_, err = vm.NewManager(*name, bridgeIfIp.String(), ipam)
	log.Println(err)

	// Clean exit.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	// Exit by user
	log.Println("Ctrl-c detected, shutting down")
	err = orchestrator.SignalOrchestratorDown(ctx, *name)
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

func initHostNetworking(ipam *vm.Ipam, outboundInterface string) error {
	// brctl addbr fcbr0
	// ip addr add 172.30.0.1/24 dev fcbr0
	// ip link set fcbr0 up
	// iptables -t nat -A POSTROUTING  -s 172.30.0.0/16 -o wlp2s0 -j MASQUERADE
	// iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
	// iptables -A FORWARD -i fcbr0 -o wlp2s0 -j ACCEPT
	bridge, err := netlink.LinkByName(vm.BridgeName)
	if err != nil {
		// if there is no bridge, let's create one
		if err.Error() == "Link not found" {
			la := netlink.NewLinkAttrs()
			la.Name = vm.BridgeName
			bridge = &netlink.Bridge{LinkAttrs: la}
			err := netlink.LinkAdd(bridge)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	ip, err := ipam.AcquireIP()
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf("Gateway IP for bridge: %s\n", ip))
	ipWithCidr := fmt.Sprintf("%s/24", ip)
	addr, err := netlink.ParseAddr(ipWithCidr)
	if err != nil {
		return err
	}
	netlink.AddrAdd(bridge, addr)
	err = netlink.LinkSetUp(bridge)

	// bridge set up, now we can setup the iptables rules
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	cidr := ipam.GetCIDR()
	err = ipt.Append("nat", "POSTROUTING", "-s", cidr, "-o", outboundInterface, "-j", "MASQUERADE")
	if err != nil {
		return err
	}
	// iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
	err = ipt.Append("filter", "FORWARD", "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	if err != nil {
		return err
	}
	// iptables -A FORWARD -i fcbr0 -o wlp2s0 -j ACCEPT
	err = ipt.Append("filter", "FORWARD", "-i", vm.BridgeName, "-o", outboundInterface, "-j", "ACCEPT")
	return err
}

func health(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "up\n")
}

func initHealthCheck(port int) {
	http.HandleFunc("/health", health)
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func getHostIP(hostInterface string) (string, error) {
	iface, err := net.InterfaceByName(hostInterface)
	if err != nil {
		return "", err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	var hostAddr net.IP
	for _, addr := range addrs {
		if hostAddr = addr.(*net.IPNet).IP.To4(); hostAddr != nil {
			break
		}
	}
	if hostAddr == nil {
		return "", fmt.Errorf("Could not get address for interface: %s", hostInterface)
	}
	return hostAddr.String(), nil
}
