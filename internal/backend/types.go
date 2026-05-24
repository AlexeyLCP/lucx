package backend

import "encoding/json"

type BackendStatus struct {
	Running bool   `json:"running"`
	Version string `json:"version"`
	PID     int    `json:"pid"`
}

type InboundSpec struct {
	Tag      string          `json:"tag"`
	Protocol string          `json:"protocol"`
	Port     int             `json:"port"`
	Listen   string          `json:"listen,omitempty"`
	Settings json.RawMessage `json:"settings"`
	Stream   json.RawMessage `json:"stream,omitempty"`
}

type InboundResult struct {
	Tag  string `json:"tag"`
	Port int    `json:"port"`
}

type OutboundSpec struct {
	Tag         string          `json:"tag"`
	Protocol    string          `json:"protocol"`
	Settings    json.RawMessage `json:"settings"`
	Stream      json.RawMessage `json:"stream,omitempty"`
	SendThrough string          `json:"sendThrough,omitempty"`
}

type OutboundResult struct {
	Tag string `json:"tag"`
}

type RoutingRule struct {
	Type        string   `json:"type"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	OutboundTag string   `json:"outboundTag"`
}

type RawConfig struct {
	Inbounds  []json.RawMessage `json:"inbounds"`
	Outbounds []json.RawMessage `json:"outbounds"`
	Routing   json.RawMessage   `json:"routing"`
}
