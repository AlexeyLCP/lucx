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

// BuildClientConfig is Task 5.
func (x *XrayBackend) BuildClientConfig(ctx context.Context, ssh backend.SSHClient, inboundTag string) (string, error) {
	return "", errors.New("not implemented")
}
