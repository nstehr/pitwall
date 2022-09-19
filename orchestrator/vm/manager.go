package vm

import (
	"context"
	"fmt"
	"log"

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
		go m.onVMStop(req.GetStop())
	}
}

func (m *Manager) onVMCreate(req *CreateVMRequest) {
	ctx := context.Background()
	vm := VM{}
	vm.Id = req.Id
	vm.ImageName = req.ImageName
	vm.Status = "BUILDING_FILESYSTEM"

	sendStatusUpdate(ctx, &vm)

	fileSystem, err := buildFilesystemFromImage(ctx, req.GetImageName())
	if err != nil {
		log.Println("Error building VM filesystem: ", err)
		vm.Status = "ERROR"
		sendStatusUpdate(ctx, &vm)
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
	}
	vmConfig.ip = ip.String()
	vmConfig.gateway = m.gateway

	tap, err := getNextTap()
	if err != nil {
		log.Println("Error aquiring TAP interface: ", err)
		vm.Status = "ERROR"
		sendStatusUpdate(ctx, &vm)
	}
	vmConfig.hostInterface = tap
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

}

func (m *Manager) onVMStop(req *StopVMRequest) {
	if machine, ok := m.virtualMachines[req.Id]; ok {
		machine.stop()
		vm := VM{}
		vm.Id = req.Id
		vm.Status = "STOPPED"
		ctx := context.Background()
		sendStatusUpdate(ctx, &vm)
	}

}

func sendStatusUpdate(ctx context.Context, vm *VM) error {
	out, err := proto.Marshal(vm)
	if err != nil {
		return err
	}

	stream.Send(ctx, "orchestrator.vm.status", out)
	if err != nil {
		return err
	}
	return nil
}
