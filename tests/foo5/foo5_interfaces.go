package foo5

import "github.com/spiral/cascade/tests/foo6"

type S5Interface struct {
}

// No deps
func (s *S5Interface) Init(fooer foo6.FooReader) error {
	fooer.Fooo()
	return nil
}



func (s *S5Interface) Configure() error {
	return nil
}

func (s *S5Interface) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *S5Interface) Close() error {
	return nil
}

func (s *S5Interface) Stop() error {
	return nil
}