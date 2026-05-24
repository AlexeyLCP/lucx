package chain

import (
	"context"
	"fmt"

	"github.com/alexeylcp/lucx-core/internal/backend"
)

type ExecutedStep struct {
	Step     Step
	Inbound  *backend.InboundResult
	Outbound *backend.OutboundResult
}

func Execute(ctx context.Context, plan *Plan, sshFactory SSHFactory) ([]ExecutedStep, error) {
	var executed []ExecutedStep
	for _, step := range plan.Steps {
		client, err := sshFactory(step.ServerID)
		if err != nil {
			rollback(ctx, executed, sshFactory)
			return nil, fmt.Errorf("step %s: ssh %s: %w", step.Operation, step.ServerID, err)
		}
		es := ExecutedStep{Step: step}
		switch step.Operation {
		case "add_inbound":
			result, err := step.Backend.AddInbound(ctx, client, *step.InboundSpec)
			if err != nil {
				client.Close()
				rollback(ctx, executed, sshFactory)
				return nil, fmt.Errorf("add_inbound on %s: %w", step.ServerID, err)
			}
			es.Inbound = &result
		case "add_outbound":
			result, err := step.Backend.AddOutbound(ctx, client, *step.OutboundSpec)
			if err != nil {
				client.Close()
				rollback(ctx, executed, sshFactory)
				return nil, fmt.Errorf("add_outbound on %s: %w", step.ServerID, err)
			}
			es.Outbound = &result
		case "set_routing":
			if err := step.Backend.SetRouting(ctx, client, step.RoutingRules); err != nil {
				client.Close()
				rollback(ctx, executed, sshFactory)
				return nil, fmt.Errorf("set_routing on %s: %w", step.ServerID, err)
			}
		}
		client.Close()
		executed = append(executed, es)
	}
	return executed, nil
}
