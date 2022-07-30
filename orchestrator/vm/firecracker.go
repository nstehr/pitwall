package vm

import (
	"context"
	"fmt"
	"log"
	"os"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
)

func startVM(ctx context.Context, opts *firecrackerOptions) (*firecracker.Machine, error) {
	// convert options to a firecracker config
	fcCfg, err := opts.getFirecrackerConfig()
	if err != nil {
		return nil, err
	}

	vmmCtx, _ := context.WithCancel(ctx)

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
		return nil, fmt.Errorf("Failed creating machine: %s", err)
	}

	if err := m.Start(vmmCtx); err != nil {
		return nil, fmt.Errorf("Failed to start machine: %v", err)
	}

	if opts.validMetadata != nil {
		m.SetMetadata(vmmCtx, opts.validMetadata)
	}

	log.Printf("Start machine was happy")
	return m, nil
}
