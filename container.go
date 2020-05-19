package cascade

type Service interface {
	Serve(upstream chan interface{}) error
	Stop()
}

type Restars interface {
	Restart(upstream chan interface{}) error
}

type Container interface {
	Service
	Register(name string, service interface{})
	Get(name string) interface{}
	Has(name string) bool
	Init() error
	List() []string
}

type ForkedContainer interface {
	Container
}

type Provider interface { // <<--
	Providers() []interface{}
}

type Register interface { // -->>
	Registers() []interface{}
}
