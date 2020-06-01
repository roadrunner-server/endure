package cascade

type (
	Service interface {
		Serve(upstream chan interface{}) error
		Stop()
	}

	Container interface {
		Service
		Register(name string, service interface{})
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
