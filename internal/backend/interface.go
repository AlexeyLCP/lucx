package backend

import "context"

type BackendType string

const (
	BackendXray      BackendType = "xray"
	BackendAWG       BackendType = "awg"
	BackendSingBox   BackendType = "sing-box"
	BackendHysteria2 BackendType = "hysteria2"
	BackendTUIC      BackendType = "tuic"
)

type SSHClient interface {
	Exec(cmd string) (string, error)
	ReadFile(path string) (string, error)
	WriteFile(path, content string) error
	Host() string
}

type ProxyBackend interface {
	Type() BackendType
	Install(ctx context.Context, ssh SSHClient) (string, error)
	Start(ctx context.Context, ssh SSHClient) error
	Stop(ctx context.Context, ssh SSHClient) error
	Status(ctx context.Context, ssh SSHClient) (BackendStatus, error)
	AddInbound(ctx context.Context, ssh SSHClient, spec InboundSpec) (InboundResult, error)
	RemoveInbound(ctx context.Context, ssh SSHClient, tag string) error
	AddOutbound(ctx context.Context, ssh SSHClient, spec OutboundSpec) (OutboundResult, error)
	RemoveOutbound(ctx context.Context, ssh SSHClient, tag string) error
	SetRouting(ctx context.Context, ssh SSHClient, rules []RoutingRule) error
	GetConfig(ctx context.Context, ssh SSHClient) (*RawConfig, error)
	BuildClientConfig(ctx context.Context, ssh SSHClient, inboundTag string) (string, error)
}
