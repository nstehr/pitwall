package vm

import (
	"context"
	"fmt"
	"log"
	"os"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
)

func startVM(ctx context.Context, opts *firecrackerOptions) error {
	// convert options to a firecracker config
	fcCfg, err := opts.getFirecrackerConfig()
	if err != nil {
		return err
	}

	vmmCtx, vmmCancel := context.WithCancel(ctx)
	defer vmmCancel()

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
		return fmt.Errorf("Failed creating machine: %s", err)
	}

	if err := m.Start(vmmCtx); err != nil {
		return fmt.Errorf("Failed to start machine: %v", err)
	}

	defer m.StopVMM()

	if opts.validMetadata != nil {
		m.SetMetadata(vmmCtx, opts.validMetadata)
	}

	// wait for the VMM to exit
	if err := m.Wait(vmmCtx); err != nil {
		return fmt.Errorf("Wait returned an error %s", err)
	}
	log.Printf("Start machine was happy")
	return nil
}
