package ssh

import "github.com/alexeylcp/lucx-core/internal/store"

// Factory creates SSH clients on-demand using server records from the store.
// Connections are never persistent — every Dial is a fresh connection.
type Factory struct {
	store *store.Store
}

// NewFactory returns a new Factory backed by the given store.
func NewFactory(s *store.Store) *Factory {
	return &Factory{store: s}
}

// Dial establishes a fresh SSH connection to the server described by srv.
// Caller MUST call Close() after use.
func (f *Factory) Dial(srv *store.Server) (*Client, error) {
	return Dial(ConnectParams{
		Host:       srv.Host,
		Port:       srv.Port,
		Username:   srv.Username,
		AuthMethod: srv.AuthMethod,
		Credential: srv.Credential,
	})
}
