package plugin6

type FooWriter interface {
	Fooo() // just stupid name
}

type S6Interface struct {
}

func (s *S6Interface) Fooo() {
	println("just FooWriter interface invoke")
}

// No deps
func (s *S6Interface) Init() error {
	return nil
}

func (s *S6Interface) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *S6Interface) Stop() error {
	return nil
}

func (s *S6Interface) Provides() []interface{} {
	return []interface{}{s.ProvideInterface}
}

func (s *S6Interface) ProvideInterface() FooWriter {
	return &S6Interface{}
}
