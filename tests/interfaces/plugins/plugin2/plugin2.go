package plugin2

type FooWriter interface {
	Fooo() // just stupid name
}

type Plugin2 struct {
}

func (s *Plugin2) Fooo() {
	println("just FooWriter interface invoke")
}

// No deps
func (s *Plugin2) Init() error {
	return nil
}

func (s *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *Plugin2) Stop() error {
	return nil
}

func (s *Plugin2) Provides() []any {
	return []any{s.ProvideInterface}
}

func (s *Plugin2) ProvideInterface() (FooWriter, error) {
	return &Plugin2{}, nil
}
