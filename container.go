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

// + Init() <<--

type Cascade struct {
	providers []Provider
	registers []Register
	services  map[string]interface{}
}

func (c *Cascade) Register(name string, svc Service) {
	c.services[name] = svc

	if r, ok := svc.(Provider); ok {
		c.providers = append(c.providers, r)
	}

	if r, ok := svc.(Register); ok {
		c.registers = append(c.registers, r)
	}
}
