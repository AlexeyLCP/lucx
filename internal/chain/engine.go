package chain

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alexeylcp/lucx-core/internal/store"
)

type Engine struct {
	store      *store.Store
	sshFactory SSHFactory
}

func NewEngine(s *store.Store, sshFactory SSHFactory) *Engine {
	return &Engine{store: s, sshFactory: sshFactory}
}

func (e *Engine) Apply(ctx context.Context, chain *store.Chain) error {
	plan, err := BuildPlan(chain)
	if err != nil {
		return fmt.Errorf("plan: %w", err)
	}
	if err := Validate(ctx, plan, e.sshFactory); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	executed, err := Execute(ctx, plan, e.sshFactory)
	if err != nil {
		e.store.UpdateChainStatus(chain.ID, "draft")
		return fmt.Errorf("execute: %w", err)
	}
	for _, es := range executed {
		var inJSON, outJSON string
		if es.Inbound != nil {
			b, _ := json.Marshal(es.Inbound)
			inJSON = string(b)
		}
		if es.Outbound != nil {
			b, _ := json.Marshal(es.Outbound)
			outJSON = string(b)
		}
		e.store.UpdateChainNodeResult(chain.ID, es.Step.NodeIndex, inJSON, outJSON)
	}
	if err := e.store.UpdateChainStatus(chain.ID, "active"); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
