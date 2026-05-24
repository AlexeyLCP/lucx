package xray

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/alexeylcp/lucx-core/internal/backend"
)

type vlessInboundSettings struct {
	Clients    []vlessClient `json:"clients"`
	Decryption string        `json:"decryption"`
}

type vlessClient struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Flow  string `json:"flow,omitempty"`
}

type streamSettings struct {
	Network         string           `json:"network"`
	Security        string           `json:"security"`
	RealitySettings *realitySettings `json:"realitySettings,omitempty"`
	TLSSettings     *tlsSettings     `json:"tlsSettings,omitempty"`
}

type realitySettings struct {
	ServerName  string `json:"serverName"`
	PublicKey   string `json:"publicKey"`
	ShortID     string `json:"shortId"`
	Fingerprint string `json:"fingerprint"`
}

type tlsSettings struct {
	ServerName string `json:"serverName"`
}

func (x *XrayBackend) BuildClientConfig(ctx context.Context, ssh backend.SSHClient, inboundTag string) (string, error) {
	rawCfg, err := x.GetConfig(ctx, ssh)
	if err != nil {
		return "", fmt.Errorf("get config: %w", err)
	}

	for _, raw := range rawCfg.Inbounds {
		var inbound struct {
			Tag      string          `json:"tag"`
			Protocol string          `json:"protocol"`
			Port     int             `json:"port"`
			Listen   string          `json:"listen"`
			Settings json.RawMessage `json:"settings"`
			Stream   json.RawMessage `json:"streamSettings"`
		}
		if err := json.Unmarshal(raw, &inbound); err != nil {
			continue
		}
		if inbound.Tag != inboundTag {
			continue
		}

		var settings vlessInboundSettings
		if err := json.Unmarshal(inbound.Settings, &settings); err != nil {
			return "", fmt.Errorf("parse settings: %w", err)
		}
		if len(settings.Clients) == 0 {
			return "", fmt.Errorf("no clients in inbound %s", inboundTag)
		}

		client := settings.Clients[0]
		host := ssh.Host()

		u := url.URL{
			Scheme: inbound.Protocol,
			User:   url.User(client.ID),
			Host:   fmt.Sprintf("%s:%d", host, inbound.Port),
		}

		q := u.Query()
		enc := settings.Decryption
		if enc == "" {
			enc = "none"
		}
		q.Set("encryption", enc)

		var stream streamSettings
		if err := json.Unmarshal(inbound.Stream, &stream); err == nil {
			if stream.Security != "" {
				q.Set("security", stream.Security)
			}
			if stream.Network != "" && stream.Network != "tcp" {
				q.Set("type", stream.Network)
			}

			if stream.RealitySettings != nil {
				rs := stream.RealitySettings
				if rs.ServerName != "" {
					q.Set("sni", rs.ServerName)
				}
				if rs.PublicKey != "" {
					q.Set("pbk", rs.PublicKey)
				}
				if rs.ShortID != "" {
					q.Set("sid", rs.ShortID)
				}
				if rs.Fingerprint != "" {
					q.Set("fp", rs.Fingerprint)
				}
			}
			if stream.TLSSettings != nil && stream.TLSSettings.ServerName != "" {
				q.Set("sni", stream.TLSSettings.ServerName)
			}
		}

		if client.Flow != "" {
			q.Set("flow", client.Flow)
		}

		u.RawQuery = q.Encode()
		return u.String() + "#LucX-" + inboundTag, nil
	}

	return "", fmt.Errorf("inbound %q not found", inboundTag)
}
