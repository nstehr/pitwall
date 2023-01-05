package orchestrator

import (
	"context"

	"github.com/nstehr/pitwall/orchestrator/stream"
	"google.golang.org/protobuf/proto"
)

var (
	healthRoutingKey = "orchestrator.health"
)

func SignalOrchestratorAlive(ctx context.Context, name string, healthCheckUrl string) error {
	orchestrator := Orchestrator{Name: name, Status: "UP", HealthCheck: healthCheckUrl}
	out, err := proto.Marshal(&orchestrator)
	if err != nil {
		return err
	}
	err = stream.Send(ctx, healthRoutingKey, out)
	return err
}

func SignalOrchestratorDown(ctx context.Context, name string) error {
	orchestrator := Orchestrator{Name: name, Status: "DOWN"}
	out, err := proto.Marshal(&orchestrator)
	if err != nil {
		return err
	}
	err = stream.Send(ctx, healthRoutingKey, out)
	return err
}
