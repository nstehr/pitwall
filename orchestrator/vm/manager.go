package vm

import (
	"context"
	"fmt"
	"log"
	"os"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/nstehr/pitwall/orchestrator/stream"
	"google.golang.org/protobuf/proto"
)

type Manager struct {
	hostname        string
	virtualMachines map[int64]*vmInstance
	ipam            *Ipam
	gateway         string
}

type vmInstance struct {
	vmConfig
	machine *firecracker.Machine
	cancel  context.CancelFunc
}

type vmConfig struct {
	kernelImagePath string
	rootFSPath      string
	ip              string
	gateway         string
	hostInterface   string
}

func (vm *vmInstance) stop() {
	vm.cancel()
	vm.machine.StopVMM()
}

func NewManager(name string, gateway string, ipam *Ipam) (*Manager, error) {
	m := &Manager{hostname: name, ipam: ipam, gateway: gateway, virtualMachines: make(map[int64]*vmInstance)}
	queue := fmt.Sprintf("orchestrator.vm.crud.%s", name)
	err := stream.RegisterHandler(queue, queue, m.dispatch)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) dispatch(msg []byte) {
	req := VMRequest{}
	err := proto.Unmarshal(msg, &req)
	if err != nil {
		log.Println("Error unmarshaling create VM message", err)
		return
	}
	switch req.Type {
	case Type_CREATE:
		go m.onVMCreate(req.GetCreate())
	case Type_DELETE:
		go m.onVMStop(req.GetStop().Id)
	}
}

func (m *Manager) onVMCreate(req *CreateVMRequest) {
	ctx := context.Background()
	vm := VM{}
	vm.Id = req.Id
	vm.ImageName = req.ImageName
	vm.Name = req.Name
	vm.Owner = req.Owner
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("error getting hostname: ", err)
	} else {
		vm.Host = hostname
	}
	vm.Status = "BUILDING_FILESYSTEM"

	sendStatusUpdate(ctx, &vm)
	fileSystem, err := buildFilesystemFromImage(ctx, req.GetImageName(), req.GetPublicKey())
	if err != nil {
		log.Println("Error building VM filesystem: ", err)
		vm.Status = "ERROR"
		sendStatusUpdate(ctx, &vm)
		return
	}

	vmConfig := vmConfig{}
	// --kernel=hello-vmlinux.bin --root-drive=hello-rootfs.ext4
	//vmConfig.rootFSPath = "hello-rootfs.ext4"
	vmConfig.kernelImagePath = "vmlinux-5.10"
	vmConfig.rootFSPath = fileSystem
	ip, err := m.ipam.AcquireIP()
	if err != nil {
		log.Println("Error aquiring IP address: ", err)
		vm.Status = "ERROR"
		sendStatusUpdate(ctx, &vm)
		return
	}
	vmConfig.ip = ip.String()
	vmConfig.gateway = m.gateway

	tap, err := getNextTap()
	if err != nil {
		log.Println("Error aquiring TAP interface: ", err)
		vm.Status = "ERROR"
		sendStatusUpdate(ctx, &vm)
		return
	}
	vmConfig.hostInterface = tap
	vm.PrivateIp = ip.String()
	vm.Status = "BOOTING"
	sendStatusUpdate(ctx, &vm)

	machine, err := startVM(ctx, vmConfig)
	if err != nil {
		log.Println("Error booting VM: ", err)
		vm.Status = "ERROR"
		sendStatusUpdate(ctx, &vm)
	}
	// TODO: make thread safe
	m.virtualMachines[req.Id] = machine
	vm.Status = "RUNNING"
	sendStatusUpdate(ctx, &vm)

	// not sure if this is the way, but Wait()
	// seems to be the way I can catch if the VM is terminated on the host
	go func() {
		ctx := context.Background()
		machine.machine.Wait(ctx)
		m.releaseResources(req.Id)
	}()

}

func (m *Manager) onVMStop(vmId int64) {
	if machine, ok := m.virtualMachines[vmId]; ok {
		machine.stop()
	}

}

func (m *Manager) releaseResources(vmId int64) {
	if machine, ok := m.virtualMachines[vmId]; ok {
		vm := VM{}
		vm.Id = vmId
		ctx := context.Background()
		err := releaseTap(machine.hostInterface)
		if err != nil {
			log.Println("Error removing tap interface: ", err)
			vm.Status = "ERROR"
			sendStatusUpdate(ctx, &vm)
			return
		}
		err = m.ipam.ReleaseIP(machine.ip)
		if err != nil {
			log.Println("Error removing releasing ip: ", err)
			vm.Status = "ERROR"
			sendStatusUpdate(ctx, &vm)
			return
		}

		vm.Status = "STOPPED"
		sendStatusUpdate(ctx, &vm)
	}
}

func sendStatusUpdate(ctx context.Context, vm *VM) error {
	out, err := proto.Marshal(vm)
	if err != nil {
		return err
	}

	stream.Send(ctx, fmt.Sprintf("orchestrator.vm.status.%s", vm.Status), out)
	if err != nil {
		return err
	}
	return nil
}
