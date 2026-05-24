package ssh

import "time"

// Dial connects to a server on-demand. Caller MUST Close() after use.
// No caching, no pooling — every call creates a fresh connection.
// This is essential for Router Mode (128-512 MB RAM).
func Dial(params ConnectParams) (*Client, error) {
	if params.Timeout == 0 {
		params.Timeout = 10 * time.Second
	}
	return Connect(params)
}
