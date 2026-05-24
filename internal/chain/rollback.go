package chain

import (
	"context"
	"log"
)

func rollback(ctx context.Context, executed []ExecutedStep, sshFactory SSHFactory) {
	log.Printf("ROLLBACK: undoing %d steps", len(executed))
	for i := len(executed) - 1; i >= 0; i-- {
		es := executed[i]
		client, err := sshFactory(es.Step.ServerID)
		if err != nil {
			log.Printf("ROLLBACK ERROR: cannot connect to %s: %v", es.Step.ServerID, err)
			continue
		}
		switch es.Step.Operation {
		case "add_inbound":
			if es.Inbound != nil {
				if err := es.Step.Backend.RemoveInbound(ctx, client, es.Inbound.Tag); err != nil {
					log.Printf("ROLLBACK ERROR: remove inbound %s: %v", es.Inbound.Tag, err)
				}
			}
		case "add_outbound":
			if es.Outbound != nil {
				if err := es.Step.Backend.RemoveOutbound(ctx, client, es.Outbound.Tag); err != nil {
					log.Printf("ROLLBACK ERROR: remove outbound %s: %v", es.Outbound.Tag, err)
				}
			}
		}
		client.Close()
	}
}
