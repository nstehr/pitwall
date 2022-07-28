package orchestrator

import (
	"context"

	"github.com/nstehr/pitwall/orchestrator/stream"
	"google.golang.org/protobuf/proto"
)

func SignalOrchestratorAlive(ctx context.Context, name string) error {
	orchestrator := Orchestrator{Name: name, Status: "UP"}
	out, err := proto.Marshal(&orchestrator)
	if err != nil {
		return err
	}
	err = stream.Send(ctx, "orchestrator.health", out)
	return err
}
