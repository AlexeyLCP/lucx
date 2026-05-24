package xray

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alexeylcp/lucx-core/internal/backend"
)

// For MVP: config.json approach (always works, gRPC not verified yet).
// Phase 0 will determine if gRPC is viable for Reality+uTLS.

func (x *XrayBackend) GetConfig(ctx context.Context, ssh backend.SSHClient) (*backend.RawConfig, error) {
	content, err := ssh.ReadFile(xrayConfigPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var raw backend.RawConfig
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &raw, nil
}

func (x *XrayBackend) AddInbound(ctx context.Context, ssh backend.SSHClient, spec backend.InboundSpec) (backend.InboundResult, error) {
	cfg, err := x.GetConfig(ctx, ssh)
	if err != nil {
		return backend.InboundResult{}, err
	}

	inbound, _ := json.Marshal(map[string]interface{}{
		"tag":      spec.Tag,
		"protocol": spec.Protocol,
		"port":     spec.Port,
		"listen":   spec.Listen,
		"settings": json.RawMessage(spec.Settings),
		"streamSettings": json.RawMessage(spec.Stream),
	})

	var inbounds []json.RawMessage
	inbounds = append(inbounds, cfg.Inbounds...)
	inbounds = append(inbounds, inbound)

	return backend.InboundResult{Tag: spec.Tag, Port: spec.Port},
		writeConfig(ctx, ssh, inbounds, cfg.Outbounds, cfg.Routing)
}

func (x *XrayBackend) RemoveInbound(ctx context.Context, ssh backend.SSHClient, tag string) error {
	cfg, err := x.GetConfig(ctx, ssh)
	if err != nil {
		return err
	}
	var filtered []json.RawMessage
	for _, raw := range cfg.Inbounds {
		var item struct{ Tag string `json:"tag"` }
		if json.Unmarshal(raw, &item) == nil && item.Tag == tag {
			continue
		}
		filtered = append(filtered, raw)
	}
	return writeConfig(ctx, ssh, filtered, cfg.Outbounds, cfg.Routing)
}

func (x *XrayBackend) AddOutbound(ctx context.Context, ssh backend.SSHClient, spec backend.OutboundSpec) (backend.OutboundResult, error) {
	cfg, err := x.GetConfig(ctx, ssh)
	if err != nil {
		return backend.OutboundResult{}, err
	}
	outbound, _ := json.Marshal(map[string]interface{}{
		"tag":      spec.Tag,
		"protocol": spec.Protocol,
		"settings": json.RawMessage(spec.Settings),
		"streamSettings": json.RawMessage(spec.Stream),
	})
	outbounds := append(cfg.Outbounds, outbound)
	return backend.OutboundResult{Tag: spec.Tag},
		writeConfig(ctx, ssh, cfg.Inbounds, outbounds, cfg.Routing)
}

func (x *XrayBackend) RemoveOutbound(ctx context.Context, ssh backend.SSHClient, tag string) error {
	cfg, err := x.GetConfig(ctx, ssh)
	if err != nil {
		return err
	}
	var filtered []json.RawMessage
	for _, raw := range cfg.Outbounds {
		var item struct{ Tag string `json:"tag"` }
		if json.Unmarshal(raw, &item) == nil && item.Tag == tag {
			continue
		}
		filtered = append(filtered, raw)
	}
	return writeConfig(ctx, ssh, cfg.Inbounds, filtered, cfg.Routing)
}

func (x *XrayBackend) SetRouting(ctx context.Context, ssh backend.SSHClient, rules []backend.RoutingRule) error {
	cfg, err := x.GetConfig(ctx, ssh)
	if err != nil {
		return err
	}
	routingJSON, _ := json.Marshal(map[string]interface{}{
		"rules": rules,
	})
	return writeConfig(ctx, ssh, cfg.Inbounds, cfg.Outbounds, routingJSON)
}

func writeConfig(ctx context.Context, ssh backend.SSHClient, inbounds, outbounds []json.RawMessage, routing json.RawMessage) error {
	cfg := map[string]interface{}{
		"inbounds":  inbounds,
		"outbounds": outbounds,
		"routing":   routing,
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := ssh.WriteFile(xrayConfigPath, string(data)); err != nil {
		return err
	}
	// Restart Xray to apply changes
	// Use SIGHUP for graceful reload if supported, otherwise restart
	ssh.Exec("pkill -HUP xray 2>/dev/null || systemctl restart xray 2>/dev/null || /etc/init.d/xray restart 2>/dev/null")
	return nil
}
