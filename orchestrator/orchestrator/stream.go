package orchestrator

import (
	"context"

	"github.com/nstehr/pitwall/orchestrator/stream"
	"google.golang.org/protobuf/proto"
)

func SignalOrchestratorAlive(name string) error {
	ctx := context.Background()
	orchestrator := Orchestrator{Name: name, Status: "UP"}
	out, err := proto.Marshal(&orchestrator)
	if err != nil {
		return err
	}
	err = stream.Send(ctx, "orchestrator.health", out)
	return err
}
