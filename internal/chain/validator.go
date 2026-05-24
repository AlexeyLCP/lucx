package chain

import (
	"context"
	"fmt"

	"github.com/alexeylcp/lucx-core/internal/ssh"
)

type SSHFactory func(serverID string) (*ssh.Client, error)

func Validate(ctx context.Context, plan *Plan, sshFactory SSHFactory) error {
	serverSet := make(map[string]bool)
	for _, step := range plan.Steps {
		serverSet[step.ServerID] = true
	}
	for serverID := range serverSet {
		client, err := sshFactory(serverID)
		if err != nil {
			return fmt.Errorf("server %s SSH: %w", serverID, err)
		}
		if _, err := client.Exec("echo ok"); err != nil {
			client.Close()
			return fmt.Errorf("server %s unreachable: %w", serverID, err)
		}
		status, err := stepForServer(plan, serverID).Backend.Status(ctx, client)
		client.Close()
		if err != nil {
			return fmt.Errorf("server %s status: %w", serverID, err)
		}
		if !status.Running {
			return fmt.Errorf("server %s: backend not running", serverID)
		}
	}
	return nil
}

func stepForServer(plan *Plan, serverID string) *Step {
	for _, s := range plan.Steps {
		if s.ServerID == serverID {
			return &s
		}
	}
	return nil
}
