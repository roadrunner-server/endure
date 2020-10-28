package ServeRetryErr

type S4 struct {
}

// No deps
func (s *S4) Init(foo5 S5) error {
	return nil
}

func (s *S4) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *S4) Stop() error {
	return nil
}
