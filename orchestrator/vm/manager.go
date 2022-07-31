package vm

import (
	"context"
	"fmt"
	"log"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/jessevdk/go-flags"
	"github.com/nstehr/pitwall/orchestrator/stream"
	"google.golang.org/protobuf/proto"
)

type Manager struct {
	hostname        string
	virtualMachines map[int64]*firecracker.Machine
}

func NewManager(name string) (*Manager, error) {
	m := &Manager{hostname: name, virtualMachines: make(map[int64]*firecracker.Machine)}
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

	buildFilesystemFromImage(ctx, req.GetImageName())
	opts := newFirecrackerOptions()
	p := flags.NewParser(opts, flags.Default)
	p.Parse()
	// --kernel=hello-vmlinux.bin --root-drive=hello-rootfs.ext4
	opts.FcKernelImage = "hello-vmlinux.bin"
	opts.FcRootDrivePath = "hello-rootfs.ext4"

	vm.Status = "BOOTING"
	sendStatusUpdate(ctx, &vm)

	machine, err := startVM(ctx, opts)
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
		machine.StopVMM()
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
