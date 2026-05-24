package scanner

import (
	"encoding/json"
	"fmt"

	"github.com/alexeylcp/lucx-core/internal/ssh"
)

// ImportedConfig holds the imported Xray configuration.
type ImportedConfig struct {
	Inbounds  []json.RawMessage `json:"inbounds"`
	Outbounds []json.RawMessage `json:"outbounds"`
	Routing   json.RawMessage   `json:"routing"`
}

// ImportExisting reads and parses an existing standalone Xray config from
// the remote server.
func ImportExisting(client *ssh.Client) (*ImportedConfig, error) {
	content, err := client.ReadFile("/usr/local/etc/xray/config.json")
	if err != nil {
		return nil, fmt.Errorf("read xray config: %w", err)
	}
	var cfg ImportedConfig
	if err := json.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("parse xray config: %w", err)
	}
	return &cfg, nil
}
