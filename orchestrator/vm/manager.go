package vm

import (
	"fmt"

	"github.com/nstehr/pitwall/orchestrator/stream"
)

type Manager struct {
	hostname string
}

func NewManager(name string) (*Manager, error) {
	queue := fmt.Sprintf("orchestrator.vm.crud.%s", name)

	err := stream.RegisterHandler(queue, queue, onVMCreate)
	if err != nil {
		return nil, err
	}
	return &Manager{hostname: name}, nil
}
