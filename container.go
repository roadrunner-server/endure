package cascade

import "context"

type (
	// TODO namings
	Graceful interface {
		// Configure is used when we need to make preparation and wait for all services till Serve
		Configure() error
		// Close frees resources allocated by the service
		Close() error
	}
	Service interface {
		// Serve
		Serve(ctx context.Context) error
	}

	Container interface {
		Service
		Register(service interface{}) error
		Get(name string) interface{}
		Has(name string) bool
		Init() error
		List() []string
	}

	// Provider declares the ability to provide service edges of declared types.
	Provider interface {
		Provides() []interface{}
	}

	// Register declares the ability to accept the plugins which match the provided method signature.
	Register interface {
		Depends() []interface{}
	}
)
