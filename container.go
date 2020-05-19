package cascade

type Container interface {
	Register(name string, service interface{})
	Get(name string) interface{}
	Has(name string) bool
	Init() error
	List() []string
}

type Runner interface {
	Container
	Serve() error
	Stop()
}

type Service interface {
	Serve(chan interface{}) error
	Stop() error
}

type Provider interface { // <<--
	Providers() []interface{}
}

type Observer interface { // -->>
	Registers() []interface{}
}

// + Init() <<--
