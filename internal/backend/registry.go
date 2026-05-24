package backend

import "fmt"

var registry = map[BackendType]func() ProxyBackend{}

func Register(t BackendType, factory func() ProxyBackend) {
	registry[t] = factory
}

func Get(t BackendType) (ProxyBackend, error) {
	factory, ok := registry[t]
	if !ok {
		return nil, fmt.Errorf("unknown backend type: %s", t)
	}
	return factory(), nil
}
