package vm

import (
	"context"

	flags "github.com/jessevdk/go-flags"
)

type ApiServer struct {
	UnimplementedVMServiceServer
}

func NewApiServer() *ApiServer {
	return &ApiServer{}
}
func (as *ApiServer) CreateVM(in *CreateVMRequest, stream VMService_CreateVMServer) error {
	ctx := context.Background()
	vm := VM{}
	vm.Status = "Pending"
	stream.Send(&vm)
	vm.Status = "Building Filesystem"
	stream.Send(&vm)
	buildFilesystemFromImage(ctx, in.GetImageName())
	go func() {
		opts := newFirecrackerOptions()
		p := flags.NewParser(opts, flags.Default)
		p.Parse()
		// --kernel=hello-vmlinux.bin --root-drive=hello-rootfs.ext4
		opts.FcKernelImage = "hello-vmlinux.bin"
		opts.FcRootDrivePath = "hello-rootfs.ext4"
		startVM(ctx, opts)
	}()
	vm.Status = "Up"
	stream.Send(&vm)
	return nil
}
