package vm

import (
	"context"
	"log"

	"github.com/jessevdk/go-flags"
	"google.golang.org/protobuf/proto"
)

func onVMCreate(msg []byte) {
	req := CreateVMRequest{}
	err := proto.Unmarshal(msg, &req)
	if err != nil {
		log.Println("Error unmarshaling create VM message", err)
		return
	}
	// after we know we can unmarshal, we can return and do the work in a goroutine
	go func() {
		ctx := context.Background()
		buildFilesystemFromImage(ctx, req.GetImageName())
		opts := newFirecrackerOptions()
		p := flags.NewParser(opts, flags.Default)
		p.Parse()
		// --kernel=hello-vmlinux.bin --root-drive=hello-rootfs.ext4
		opts.FcKernelImage = "hello-vmlinux.bin"
		opts.FcRootDrivePath = "hello-rootfs.ext4"
		startVM(ctx, opts)
	}()
}
