package foo6

type FooReader interface {
	Fooo() // just stupid name
}

type S6Interface struct {
}

func (s *S6Interface) Fooo() {
	println("bueeeeeeeee")
}

// No deps
func (s *S6Interface) Init() error {
	return nil
}

func (s *S6Interface) Configure() error {
	return nil
}

func (s *S6Interface) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *S6Interface) Close() error {
	return nil
}

func (s *S6Interface) Stop() error {
	return nil
}

func (s *S6Interface) Provides() []interface{} {
	return []interface{}{s.ProvideInterface}
}

func (s *S6Interface) ProvideInterface() (FooReader, error) {
	return &S6Interface{}, nil
}
