package vm

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

const (
	rwDeviceSuffix = ":rw"
	roDeviceSuffix = ":ro"
)

func startVM(ctx context.Context, cfg vmConfig) (*vmInstance, error) {

	// set up VM config
	fcCfg := firecracker.Config{}
	fcCfg.SocketPath = getSocketPath()
	fcCfg.KernelImagePath = cfg.kernelImagePath
	fcCfg.KernelArgs = "ro console=ttyS0 noapic reboot=k panic=1 pci=off nomodules"

	blockDevices, err := getRootDrive(cfg.rootFSPath)
	if err != nil {
		return nil, err
	}

	fcCfg.Drives = blockDevices

	fcCfg.MachineCfg = models.MachineConfiguration{
		VcpuCount:  firecracker.Int64(1),
		Smt:        firecracker.Bool(true),
		MemSizeMib: firecracker.Int64(512),
	}

	// ip addr add 172.16.0.2/24 dev eth0
	// ip link set eth0 up
	// ip route add default via 172.16.0.1 dev eth0

	netConfig := firecracker.StaticNetworkConfiguration{}
	netConfig.HostDevName = cfg.hostInterface
	netConfig.MacAddress = "AA:FC:00:00:00:01"
	ipConfig := firecracker.IPConfiguration{}
	ipConfig.Gateway = net.ParseIP(cfg.gateway)
	ipNet := net.IPNet{}
	ipNet.IP = net.ParseIP(cfg.ip)
	ipNet.Mask = net.IPv4Mask(255, 255, 255, 0)
	ipConfig.IPAddr = ipNet
	ipConfig.IfName = "eth0"
	ipConfig.Nameservers = []string{"8.8.8.8"}
	netConfig.IPConfiguration = &ipConfig
	ni := firecracker.NetworkInterface{StaticConfiguration: &netConfig}
	fcCfg.NetworkInterfaces = []firecracker.NetworkInterface{ni}

	fcCfg.LogLevel = "DEBUG"

	vmmCtx, cancelFunc := context.WithCancel(ctx)

	// machineOpts := []firecracker.Opt{
	// 	firecracker.WithLogger(log.NewEntry(logger)),
	// }

	machineOpts := []firecracker.Opt{}

	cmd := firecracker.VMCommandBuilder{}.
		WithSocketPath(fcCfg.SocketPath).
		WithStdin(os.Stdin).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		Build(ctx)

	machineOpts = append(machineOpts, firecracker.WithProcessRunner(cmd))

	m, err := firecracker.NewMachine(vmmCtx, fcCfg, machineOpts...)
	if err != nil {
		cancelFunc()
		return nil, fmt.Errorf("Failed creating machine: %s", err)
	}

	if err := m.Start(vmmCtx); err != nil {
		cancelFunc()
		return nil, fmt.Errorf("Failed to start machine: %v", err)
	}

	log.Printf("Start machine was happy")
	vmInstance := &vmInstance{machine: m, cancel: cancelFunc, vmConfig: cfg}
	return vmInstance, nil
}

func getRootDrive(rootDrivePath string) ([]models.Drive, error) {
	rootDrivePath, readOnly := parseDevice(rootDrivePath)
	rootDrive := models.Drive{
		DriveID:      firecracker.String("1"),
		PathOnHost:   firecracker.String(rootDrivePath),
		IsReadOnly:   firecracker.Bool(readOnly),
		IsRootDevice: firecracker.Bool(true),
	}
	return []models.Drive{rootDrive}, nil
}

// following functions from: https://github.com/firecracker-microvm/firectl/blob/ec72798240c0561dea8341d828e8c72bb0cc36c5/options.go

// Given a string in the form of path:suffix return the path and read-only marker
func parseDevice(entry string) (path string, readOnly bool) {
	if strings.HasSuffix(entry, roDeviceSuffix) {
		return strings.TrimSuffix(entry, roDeviceSuffix), true
	}

	return strings.TrimSuffix(entry, rwDeviceSuffix), false
}

// getSocketPath provides a randomized socket path by building a unique filename
// and searching for the existence of directories {$HOME, os.TempDir()} and returning
// the path with the first directory joined with the unique filename. If we can't
// find a good path panics.
func getSocketPath() string {
	filename := strings.Join([]string{
		".firecracker.sock",
		strconv.Itoa(os.Getpid()),
		strconv.Itoa(rand.Intn(1000))},
		"-",
	)
	var dir string
	if d := os.Getenv("HOME"); checkExistsAndDir(d) {
		dir = d
	} else if checkExistsAndDir(os.TempDir()) {
		dir = os.TempDir()
	} else {
		panic("Unable to find a location for firecracker socket.")
	}

	return filepath.Join(dir, filename)
}

// checkExistsAndDir returns true if path exists and is a Dir
func checkExistsAndDir(path string) bool {
	// empty
	if path == "" {
		return false
	}
	// does it exist?
	if info, err := os.Stat(path); err == nil {
		// is it a directory?
		return info.IsDir()
	}
	return false
}
