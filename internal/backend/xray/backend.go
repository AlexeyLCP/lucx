package xray

import (
	"github.com/alexeylcp/lucx-core/internal/backend"
)

type XrayBackend struct{}

func init() {
	backend.Register(backend.BackendXray, func() backend.ProxyBackend {
		return &XrayBackend{}
	})
}

func (x *XrayBackend) Type() backend.BackendType { return backend.BackendXray }
