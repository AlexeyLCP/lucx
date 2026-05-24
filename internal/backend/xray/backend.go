package xray

import (
	"context"
	"errors"

	"github.com/alexeylcp/lucx-core/internal/backend"
)

type XrayBackend struct{}

func init() {
	backend.Register(backend.BackendXray, func() backend.ProxyBackend {
		return &XrayBackend{}
	})
}

func (x *XrayBackend) Type() backend.BackendType { return backend.BackendXray }

func (x *XrayBackend) Install(ctx context.Context, ssh backend.SSHClient) (string, error) {
	return "", errors.New("not implemented")
}
func (x *XrayBackend) Start(ctx context.Context, ssh backend.SSHClient) error {
	return errors.New("not implemented")
}
func (x *XrayBackend) Stop(ctx context.Context, ssh backend.SSHClient) error {
	return errors.New("not implemented")
}
func (x *XrayBackend) Status(ctx context.Context, ssh backend.SSHClient) (backend.BackendStatus, error) {
	return backend.BackendStatus{}, errors.New("not implemented")
}
func (x *XrayBackend) AddInbound(ctx context.Context, ssh backend.SSHClient, spec backend.InboundSpec) (backend.InboundResult, error) {
	return backend.InboundResult{}, errors.New("not implemented")
}
func (x *XrayBackend) RemoveInbound(ctx context.Context, ssh backend.SSHClient, tag string) error {
	return errors.New("not implemented")
}
func (x *XrayBackend) AddOutbound(ctx context.Context, ssh backend.SSHClient, spec backend.OutboundSpec) (backend.OutboundResult, error) {
	return backend.OutboundResult{}, errors.New("not implemented")
}
func (x *XrayBackend) RemoveOutbound(ctx context.Context, ssh backend.SSHClient, tag string) error {
	return errors.New("not implemented")
}
func (x *XrayBackend) SetRouting(ctx context.Context, ssh backend.SSHClient, rules []backend.RoutingRule) error {
	return errors.New("not implemented")
}
func (x *XrayBackend) GetConfig(ctx context.Context, ssh backend.SSHClient) (*backend.RawConfig, error) {
	return nil, errors.New("not implemented")
}
func (x *XrayBackend) BuildClientConfig(ctx context.Context, ssh backend.SSHClient, inboundTag string) (string, error) {
	return "", errors.New("not implemented")
}
