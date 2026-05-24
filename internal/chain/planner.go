package chain

import (
	"fmt"

	"github.com/alexeylcp/lucx-core/internal/backend"
	"github.com/alexeylcp/lucx-core/internal/store"
)

type Step struct {
	NodeIndex    int
	Operation    string
	ServerID     string
	Backend      backend.ProxyBackend
	InboundSpec  *backend.InboundSpec
	OutboundSpec *backend.OutboundSpec
	RoutingRules []backend.RoutingRule
}

type Plan struct {
	Steps []Step
}

func BuildPlan(chain *store.Chain) (*Plan, error) {
	plan := &Plan{}
	chainID := chain.ID

	for i, node := range chain.Nodes {
		be, err := backend.Get(backend.BackendType(node.BackendType))
		if err != nil {
			return nil, fmt.Errorf("node %d: %w", i, err)
		}

		if node.Role == "entry" {
			plan.Steps = append(plan.Steps, Step{
				NodeIndex: i,
				Operation: "add_inbound",
				ServerID:  node.ServerID,
				Backend:   be,
				InboundSpec: &backend.InboundSpec{
					Tag:      fmt.Sprintf("lucx-%s-entry", chainID),
					Protocol: node.Protocol,
					Port:     443,
					Listen:   "0.0.0.0",
				},
			})
		}

		if node.Role != "exit" && i+1 < len(chain.Nodes) {
			outTag := fmt.Sprintf("lucx-%s-hop%d-to-%d", chainID, i, i+1)
			plan.Steps = append(plan.Steps, Step{
				NodeIndex: i,
				Operation: "add_outbound",
				ServerID:  node.ServerID,
				Backend:   be,
				OutboundSpec: &backend.OutboundSpec{
					Tag:      outTag,
					Protocol: node.Protocol,
				},
			})
			inTag := fmt.Sprintf("lucx-%s-entry", chainID)
			if node.Role == "hop" {
				inTag = fmt.Sprintf("lucx-%s-hop%d", chainID, i)
			}
			plan.Steps = append(plan.Steps, Step{
				NodeIndex: i,
				Operation: "set_routing",
				ServerID:  node.ServerID,
				Backend:   be,
				RoutingRules: []backend.RoutingRule{
					{Type: "field", InboundTag: []string{inTag}, OutboundTag: outTag},
				},
			})
		}

		if node.Role == "hop" {
			plan.Steps = append(plan.Steps, Step{
				NodeIndex: i,
				Operation: "add_inbound",
				ServerID:  node.ServerID,
				Backend:   be,
				InboundSpec: &backend.InboundSpec{
					Tag:      fmt.Sprintf("lucx-%s-hop%d", chainID, i),
					Protocol: node.Protocol,
					Port:     443,
					Listen:   "0.0.0.0",
				},
			})
		}

		if node.Role == "exit" && i != 0 {
			plan.Steps = append(plan.Steps, Step{
				NodeIndex: i,
				Operation: "add_inbound",
				ServerID:  node.ServerID,
				Backend:   be,
				InboundSpec: &backend.InboundSpec{
					Tag:      fmt.Sprintf("lucx-%s-exit", chainID),
					Protocol: node.Protocol,
					Port:     443,
					Listen:   "0.0.0.0",
				},
			})
		}
	}
	return plan, nil
}
